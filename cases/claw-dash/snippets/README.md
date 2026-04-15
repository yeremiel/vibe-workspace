# 코드 스니펫

전체 코드는 비공개이며, 기능 설명을 위해 핵심 로직을 발췌했다.

> `.go.txt` 확장자는 Go 툴체인이 이 디렉터리를 패키지로 인식하지 않도록 `.txt`를 붙인 것이다.
> 실제 Go 코드이며, GitHub에서 구문 강조 없이 표시된다.

| 파일 | 내용 |
|------|------|
| [gateway-client.go.txt](./gateway-client.go.txt) | WebSocket Gateway 클라이언트 — challenge 핸드셰이크, exponential backoff 재연결, 동시성 안전 RPC |
| [sse-handler.go.txt](./sse-handler.go.txt) | Gin 기반 SSE 스트리밍 핸들러 — 이벤트 구독, keepalive |
| [agent-board-dashboard.svelte](./agent-board-dashboard.svelte) | Svelte 5 Runes 칸반 보드 컴포넌트 — `$props()`, `$state()`, `$derived()` 패턴 |
