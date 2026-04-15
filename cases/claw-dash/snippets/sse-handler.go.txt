//go:build ignore
// This file is an excerpt for documentation purposes. Not meant to be compiled standalone.

// 발췌 출처: claw-dash backend/internal/handler/events/events_api.go
// 전체 코드는 비공개입니다. 핵심 로직 발췌.
//
// Go Gin 기반 SSE(Server-Sent Events) 스트리밍 핸들러.
// - Gateway 이벤트를 채널로 구독하고 클라이언트로 실시간 스트리밍
// - 15초 keepalive 코멘트로 프록시/브라우저 타임아웃 방지
// - 이벤트 히스토리 조회 API (cursor 기반 페이지네이션)

package events

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"claw-dash/backend/internal/gateway"
)

// Events는 Gateway 이벤트를 SSE로 브라우저에 스트리밍한다.
// 클라이언트 연결 해제(context cancel) 시 자동으로 구독을 해제한다.
func Events(client *gateway.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		stream, unsubscribe := client.SubscribeEvents(64)
		defer unsubscribe()

		headers := c.Writer.Header()
		headers.Set("Content-Type", "text/event-stream")
		headers.Set("Cache-Control", "no-cache")
		headers.Set("Connection", "keep-alive")
		headers.Set("X-Accel-Buffering", "no") // nginx 버퍼링 비활성화

		flusher, ok := c.Writer.(http.Flusher)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "streaming unsupported"})
			return
		}

		c.Status(http.StatusOK)
		flusher.Flush()

		keepaliveTicker := time.NewTicker(15 * time.Second)
		defer keepaliveTicker.Stop()

		for {
			select {
			case <-c.Request.Context().Done():
				// 클라이언트 연결 해제 또는 서버 종료
				return
			case record, ok := <-stream:
				if !ok {
					return
				}

				normalized, ok := normalizeGatewayEvent(record)
				if !ok {
					continue
				}

				encoded, err := json.Marshal(normalized)
				if err != nil {
					continue
				}

				// SSE 표준 형식: "data: <json>\n\n"
				if _, err := fmt.Fprintf(c.Writer, "data: %s\n\n", encoded); err != nil {
					return
				}
				flusher.Flush()

			case <-keepaliveTicker.C:
				// SSE 코멘트 라인으로 keepalive 전송 (클라이언트는 무시)
				if _, err := io.WriteString(c.Writer, ": keepalive\n\n"); err != nil {
					return
				}
				flusher.Flush()
			}
		}
	}
}

// EventsHistory는 메모리 내 이벤트 히스토리를 cursor 기반으로 조회한다.
// sessionKey 필터와 before(cursor) 파라미터를 지원한다.
func EventsHistory(client *gateway.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionKey := strings.TrimSpace(c.Query("sessionKey"))
		beforeID := parseBeforeID(c.Query("before"))
		limit := parseHistoryLimit(c.Query("limit"))

		records, hasMore := client.EventHistory(limit, beforeID, sessionKey)
		items := make([]dashboardEvent, 0, len(records))

		for _, record := range records {
			normalized, ok := normalizeGatewayEvent(record)
			if !ok {
				continue
			}
			items = append(items, normalized)
		}

		response := eventsHistoryResponse{
			Items:   items,
			HasMore: hasMore,
		}

		if hasMore && len(items) > 0 {
			response.NextBefore = items[len(items)-1].ID
		}

		c.JSON(http.StatusOK, response)
	}
}

// --- 응답 타입 ---

const (
	defaultEventsHistoryLimit = 200
	maxEventsHistoryLimit     = 500
)

type dashboardEvent struct {
	ID         string `json:"id"`
	Type       string `json:"type"`
	Event      string `json:"event"`
	Payload    any    `json:"payload"`
	ReceivedAt string `json:"receivedAt"`
}

type eventsHistoryResponse struct {
	Items      []dashboardEvent `json:"items"`
	HasMore    bool             `json:"hasMore"`
	NextBefore string           `json:"nextBefore,omitempty"`
}
