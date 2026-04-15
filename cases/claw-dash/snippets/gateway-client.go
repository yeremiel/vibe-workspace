// 발췌 출처: claw-dash backend/internal/gateway/client.go
// 전체 코드는 비공개입니다. 핵심 로직 발췌.
//
// OpenClaw Gateway WebSocket 연결 클라이언트.
// - challenge → connect 핸드셰이크 수행 (protocolVersion 협상, deviceToken 갱신)
// - exponential backoff 자동 재연결 루프
// - 동시성 안전 RPC 호출 (pending map + channel 기반)
// - 연결 중단 시 모든 pending 요청 즉시 실패 처리

package gateway

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

var ErrDisconnected = errors.New("gateway disconnected")

const (
	protocolVersion       = 3
	gatewayClientID       = "gateway-client"
	gatewayClientMode     = "backend"
	defaultClientVersion  = "claw-dash-mvp"
	defaultGatewayRole    = "operator"
	connectChallengeEvent = "connect.challenge"
)

type RPCError struct {
	Code    string          `json:"code"`
	Message string          `json:"message"`
	Details json.RawMessage `json:"details,omitempty"`
}

func (e *RPCError) Error() string {
	if e.Code == "" {
		return fmt.Sprintf("gateway rpc error: %s", e.Message)
	}
	return fmt.Sprintf("gateway rpc error (%s): %s", e.Code, e.Message)
}

// Client는 OpenClaw Gateway와의 WebSocket 연결을 관리한다.
// Start() 호출 후 내부 goroutine이 연결·재연결·읽기 루프를 담당하며,
// Call()로 RPC 요청, SubscribeEvents()로 게이트웨이 이벤트 구독이 가능하다.
type Client struct {
	gatewayURL   string
	gatewayToken string
	logger       *slog.Logger
	device       *deviceIdentity

	dialer *websocket.Dialer
	events *EventHub

	reconnectMin      time.Duration
	reconnectMax      time.Duration
	handshakeTimeout  time.Duration
	connectScopes     []string
	connectClientName string

	connMu    sync.RWMutex
	conn      *websocket.Conn
	connected bool
	writeMu   sync.Mutex

	pendingMu sync.Mutex
	pending   map[string]chan pendingResponse

	nextRequestID atomic.Uint64
	startOnce     sync.Once
	closeOnce     sync.Once
	closeCh       chan struct{}
}

// --- 프레임 타입 정의 ---

type requestFrame struct {
	Type   string `json:"type"`
	ID     string `json:"id"`
	Method string `json:"method"`
	Params any    `json:"params,omitempty"`
}

type responseFrame struct {
	Type    string          `json:"type"`
	ID      string          `json:"id"`
	OK      bool            `json:"ok"`
	Payload json.RawMessage `json:"payload,omitempty"`
	Error   *frameError     `json:"error,omitempty"`
}

type eventFrame struct {
	Type    string          `json:"type"`
	Event   string          `json:"event"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

type frameError struct {
	Code    string          `json:"code"`
	Message string          `json:"message"`
	Details json.RawMessage `json:"details,omitempty"`
}

type pendingResponse struct {
	payload json.RawMessage
	rpcErr  *RPCError
	err     error
}

func NewClient(gatewayURL, gatewayToken, deviceIdentityPath string, logger *slog.Logger) *Client {
	if logger == nil {
		logger = slog.Default()
	}

	device, err := loadOrCreateDeviceIdentity(deviceIdentityPath)
	if err != nil {
		logger.Warn("failed to initialize gateway device identity", "error", err)
	}

	return &Client{
		gatewayURL:       gatewayURL,
		gatewayToken:     gatewayToken,
		logger:           logger,
		device:           device,
		dialer:           websocket.DefaultDialer,
		events:           NewEventHub(),
		reconnectMin:     time.Second,
		reconnectMax:     30 * time.Second,
		handshakeTimeout: 10 * time.Second,
		connectScopes: []string{
			"operator.read",
			"operator.admin",
		},
		connectClientName: gatewayClientID,
		pending:           make(map[string]chan pendingResponse),
		closeCh:           make(chan struct{}),
	}
}

func (c *Client) Start(ctx context.Context) {
	c.startOnce.Do(func() {
		go c.connectionLoop(ctx)
	})
}

func (c *Client) Close() error {
	var closeErr error
	c.closeOnce.Do(func() {
		close(c.closeCh)
		closeErr = c.closeConnection()
		c.failAllPending(ErrDisconnected)
	})
	return closeErr
}

func (c *Client) IsConnected() bool {
	c.connMu.RLock()
	defer c.connMu.RUnlock()
	return c.connected && c.conn != nil
}

func (c *Client) SubscribeEvents(buffer int) (<-chan EventRecord, func()) {
	return c.events.Subscribe(buffer)
}

func (c *Client) EventHistory(limit int, beforeID int64, sessionKey string) ([]EventRecord, bool) {
	return c.events.QueryHistory(HistoryQuery{
		Limit:      limit,
		BeforeID:   beforeID,
		SessionKey: sessionKey,
	})
}

// Call은 Gateway로 RPC 요청을 보내고 응답을 기다린다.
// 연결이 끊어지면 즉시 ErrDisconnected를 반환한다.
func (c *Client) Call(ctx context.Context, method string, params any) (json.RawMessage, error) {
	if method == "" {
		return nil, errors.New("rpc method is required")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	requestID := c.nextID()
	responseCh := make(chan pendingResponse, 1)

	c.pendingMu.Lock()
	c.pending[requestID] = responseCh
	c.pendingMu.Unlock()

	request := requestFrame{
		Type:   "req",
		ID:     requestID,
		Method: method,
		Params: params,
	}

	if err := c.writeFrame(request); err != nil {
		c.removePending(requestID)
		return nil, err
	}

	select {
	case <-ctx.Done():
		c.removePending(requestID)
		return nil, ctx.Err()
	case response, ok := <-responseCh:
		if !ok {
			return nil, ErrDisconnected
		}
		if response.err != nil {
			return nil, response.err
		}
		if response.rpcErr != nil {
			return nil, response.rpcErr
		}
		return response.payload, nil
	}
}

// connectionLoop은 연결 실패 시 exponential backoff로 재시도한다.
// 최소 1초, 최대 30초 간격으로 재연결을 시도하며,
// 성공 시 인터벌을 reconnectMin으로 리셋한다.
func (c *Client) connectionLoop(ctx context.Context) {
	retry := c.reconnectMin

	for {
		if c.isClosed() || ctx.Err() != nil {
			_ = c.closeConnection()
			return
		}

		conn, err := c.connect(ctx)
		if err != nil {
			c.logger.Warn("gateway dial failed", "url", c.gatewayURL, "error", err, "retry_in", retry.String())
			if !c.wait(ctx, retry) {
				return
			}
			retry = minDuration(retry*2, c.reconnectMax)
			continue
		}

		if err := c.performHandshake(ctx, conn); err != nil {
			c.logger.Warn("gateway handshake failed", "url", c.gatewayURL, "error", err, "retry_in", retry.String())
			_ = conn.Close()
			if !c.wait(ctx, retry) {
				return
			}
			retry = minDuration(retry*2, c.reconnectMax)
			continue
		}

		retry = c.reconnectMin
		c.setConnection(conn, true)
		c.logger.Info("gateway connected", "url", c.gatewayURL)

		readErr := c.readLoop(ctx, conn)
		if readErr != nil && !errors.Is(readErr, context.Canceled) && !c.isClosed() {
			c.logger.Warn("gateway connection closed", "error", readErr)
		}

		c.clearConnection(conn, false)
		c.failAllPending(ErrDisconnected)

		if !c.wait(ctx, retry) {
			return
		}
		retry = minDuration(retry*2, c.reconnectMax)
	}
}

func (c *Client) connect(ctx context.Context) (*websocket.Conn, error) {
	headers := make(http.Header)
	if c.gatewayURL != "" {
		// 실제 토큰 값은 환경 변수(OPENCLAW_GATEWAY_TOKEN)로 주입
		headers.Set("Authorization", "Bearer [REDACTED]")
	}

	conn, resp, err := c.dialer.DialContext(ctx, c.gatewayURL, headers)
	if err != nil {
		if resp != nil {
			return nil, fmt.Errorf("dial status %d: %w", resp.StatusCode, err)
		}
		return nil, err
	}

	return conn, nil
}

func (c *Client) readLoop(ctx context.Context, conn *websocket.Conn) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-c.closeCh:
			return nil
		default:
		}

		_, message, err := conn.ReadMessage()
		if err != nil {
			return err
		}

		c.handleFrame(message)
	}
}

// performHandshake는 connect.challenge 이벤트를 수신하고
// connect RPC 요청으로 응답하는 핸드셰이크 시퀀스를 수행한다.
// deviceToken이 있으면 재사용하고, token mismatch 오류 시 stale 토큰을 제거한다.
func (c *Client) performHandshake(ctx context.Context, conn *websocket.Conn) error {
	timer := time.NewTimer(c.handshakeTimeout)
	defer timer.Stop()

	connectRequestID := ""
	connectRole := defaultGatewayRole
	usedStoredDeviceToken := false

	for {
		type readResult struct {
			message []byte
			err     error
		}

		ch := make(chan readResult, 1)
		go func() {
			_, message, err := conn.ReadMessage()
			ch <- readResult{message: message, err: err}
		}()

		var message []byte
		select {
		case <-ctx.Done():
			_ = conn.Close()
			return ctx.Err()
		case <-c.closeCh:
			_ = conn.Close()
			return ErrDisconnected
		case <-timer.C:
			_ = conn.Close()
			return errors.New("gateway connect challenge timeout")
		case result := <-ch:
			if result.err != nil {
				return result.err
			}
			message = result.message
		}

		var base struct {
			Type string `json:"type"`
		}
		if err := json.Unmarshal(message, &base); err != nil {
			continue
		}

		switch base.Type {
		case "event":
			var event eventFrame
			if err := json.Unmarshal(message, &event); err != nil {
				continue
			}

			if event.Event != connectChallengeEvent || connectRequestID != "" {
				continue
			}

			var challenge struct {
				Nonce string `json:"nonce"`
			}
			if err := json.Unmarshal(event.Payload, &challenge); err != nil {
				c.logger.Debug("failed to decode gateway connect challenge payload", "error", err)
			}

			connectRequestID = c.nextID()
			connectParams, authChoice := c.buildConnectParams(challenge.Nonce)
			connectRole = connectParams.Role
			usedStoredDeviceToken = authChoice.usingStoredDeviceToken
			connectRequest := requestFrame{
				Type:   "req",
				ID:     connectRequestID,
				Method: "connect",
				Params: connectParams,
			}

			if err := c.writeJSON(conn, connectRequest); err != nil {
				return err
			}

		case "res":
			if connectRequestID == "" {
				continue
			}

			var response responseFrame
			if err := json.Unmarshal(message, &response); err != nil {
				continue
			}

			if response.ID != connectRequestID {
				continue
			}

			if !response.OK {
				if response.Error != nil {
					c.handleConnectTokenMismatch(connectRole, usedStoredDeviceToken, response.Error)
					return &RPCError{
						Code:    response.Error.Code,
						Message: response.Error.Message,
						Details: response.Error.Details,
					}
				}
				return errors.New("gateway connect failed")
			}

			c.persistConnectDeviceToken(connectRole, response.Payload)
			return nil
		}
	}
}

func (c *Client) handleFrame(message []byte) {
	var base struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(message, &base); err != nil {
		c.logger.Debug("failed to decode gateway payload", "error", err)
		return
	}

	switch base.Type {
	case "res":
		c.handleResponseFrame(message)
	case "event":
		c.handleEventFrame(message)
	}
}

func (c *Client) handleEventFrame(message []byte) {
	var event eventFrame
	if err := json.Unmarshal(message, &event); err != nil {
		return
	}

	// connect.challenge는 핸드셰이크 전용 — 앱 이벤트로 전달하지 않는다.
	if event.Event == connectChallengeEvent {
		return
	}

	c.events.Publish(message)
}

func (c *Client) failAllPending(err error) {
	c.pendingMu.Lock()
	pending := c.pending
	c.pending = make(map[string]chan pendingResponse)
	c.pendingMu.Unlock()

	for _, ch := range pending {
		select {
		case ch <- pendingResponse{err: err}:
		default:
		}
		close(ch)
	}
}

func minDuration(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}

// buildConnectParams, selectConnectAuth, persistConnectDeviceToken 등
// 인증 파라미터 조립 로직은 생략 (디바이스 토큰 관리 포함)
