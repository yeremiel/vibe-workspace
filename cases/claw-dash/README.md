# OpenClaw Workbench — 케이스 스터디

[🇯🇵 日本語](README.ja.md) | 🇰🇷 한국어

> AI 에이전트 런타임(OpenClaw Gateway)을 운영하기 위한 대시보드 겸 오케스트레이션 도구.
> Go BFF + Svelte 5 SPA 구성, MVP부터 칸반 기반 에이전트 작업 보드까지 7개 Phase를 완주했다.

---

## 1. 배경 — "왜 이걸 만들었나?"

OpenClaw Gateway는 AI 에이전트를 실행하는 런타임이지만, 세션 상태를 한눈에 보거나 에이전트에게 지시를 내리려면 매번 CLI와 JSON-RPC를 직접 다뤄야 했다. 세션이 늘어날수록 터미널 여러 개를 띄우고 수동으로 상태를 확인하는 방식은 한계가 명확했다.

필요한 건 세 가지였다. 첫째, 에이전트 세션을 실시간으로 모니터링할 수 있는 대시보드. 둘째, 채팅으로 에이전트에게 지시를 보내고 응답을 받는 인터페이스. 셋째, 여러 에이전트 작업을 칸반 보드로 오케스트레이션하는 도구. 이 세 가지를 하나의 웹 앱으로 만든 것이 OpenClaw Workbench다.

---

## 2. AI 컨텍스트 설계 — "AI에게 어떻게 일을 시켰나?"

이 프로젝트는 Claude Code를 메인 구현 도구로 사용했다. AI가 프로젝트 맥락을 정확히 파악하도록 CLAUDE.md를 계층화한 것이 핵심이다.

- **워크스페이스 CLAUDE.md**: 작업자 정보, 디렉터리 구조, 공통 금지사항, 언어 규칙 등 모든 프로젝트에 적용되는 기반 컨텍스트
- **knowledge_base/go/**: Go 코딩 컨벤션, Clean Architecture 패턴 문서. `@` 참조로 필요할 때만 로드해서 컨텍스트 낭비를 줄임
- **claw-dash/CLAUDE.md**: 프로젝트 특화 규칙. BFF 특성상 Response Envelope 미적용, Gateway 장애 시 HTTP 200 fallback 같은 표준 예외를 명시 ([CLAUDE.md 참조](./CLAUDE.md))

작업 사이클은 **설계 문서 작성 -> 구현 플랜 생성 -> 단계별 실행**을 반복했다. 백엔드 구조 리팩터링처럼 영향 범위가 큰 작업은 반드시 디자인 문서를 먼저 만들고 승인한 뒤 실행했다 ([backend-structure-refactor-design](./plans/2026-03-20-backend-structure-refactor-design.md)). 칸반 안정화처럼 복잡한 도메인은 claim matrix로 현재 상태를 사실/가설로 분류하고, phase별 scope lock을 걸어 의미 변경과 구조 변경이 섞이지 않도록 제어했다 ([kanban-stabilization-plan](./plans/2026-03-24-kanban-stabilization-plan.md)). 프로젝트 전체에서 50개 이상의 플랜 문서가 생성됐다.

---

## 3. 개발 흐름 — "실제로 어떻게 진행했나?"

### 기술 선택

**BFF를 Node 대신 Go로 선택했다.** myLib 백엔드에서 Go/Gin을 이미 사용하고 있었고, BFF의 핵심 역할인 WebSocket 연결 관리와 SSE fan-out에서 goroutine이 Node의 event loop보다 직관적이었다. 처음에 Node로 시작하고 나중에 Go로 재작성하면 같은 API를 두 번 구현하는 셈이므로, 처음부터 Go를 선택했다.

**프론트엔드를 React 대신 Svelte 5로 선택했다.** Angular와 JSP 경험자 입장에서 Svelte의 HTML-first 템플릿이 React JSX보다 자연스러웠다. `$state`, `$derived`, `$effect` Runes로 상태를 관리하면 useState/useEffect 훅의 클로저 함정이 없다. 무엇보다 안 써본 기술을 써보는 것 자체가 개인 프로젝트의 목적이었다 ([agent-board-dashboard.svelte 참조](./snippets/agent-board-dashboard.svelte)).

### 아키텍처 진화

초기 MVP에서는 `cmd/server/main.go`에 의존성 조립과 라우팅이 집중되어 있었다. 기능이 빠르게 추가되면서 핸들러도 단일 디렉토리에 누적됐고, 비즈니스 로직은 `internal/kanban`, `internal/usage`, `internal/gateway`로 흩어져 책임 경계가 모호해졌다.

이를 `Handler -> Service -> Repository` 계층으로 리팩터링했다. main.go를 얇게 유지하고 앱 조립 책임을 `internal/app`으로 이동시켰고, 도메인별로 `router`, `handler`, `service`를 분리했다. 기존 API 경로와 JSON 응답 shape는 변경하지 않는 것을 원칙으로 삼았다 ([backend-structure-refactor-design 참조](./plans/2026-03-20-backend-structure-refactor-design.md)).

### 서버 권위 큐 도입

Chat 입력 처리에서 중요한 설계 결정이 있었다. 프론트 상태만으로 전송을 제어하면, 응답 전 연속 입력 시 turn 경합이 발생하고, Stop 직후 재전송에서 race condition이 생기며, 새로고침이나 다중 클라이언트 환경에서 상태가 불일치한다.

이를 해결하기 위해 세션별 단일 in-flight + FIFO 큐를 서버에서 관리하도록 했다. `/api/sessions/send`는 `accepted` 또는 `queued + queueIndex`를 명시적으로 반환해서, 프론트엔드가 현재 메시지의 처리 상태를 정확히 알 수 있게 했다 ([session-chat-inline-message-design 참조](./plans/2026-03-19-session-chat-inline-message-design.md)).

---

## 4. 좌충우돌 — "막혔을 때 어떻게 했나?"

### 백엔드 재시작 실패 시 구버전 프로세스가 계속 동작

코드를 수정하고 백엔드를 재시작했는데 동작이 이전과 동일했다. 새 프로세스가 포트 바인딩에 실패해 즉시 종료했지만, 구버전 서버가 살아 있어서 헬스체크가 정상처럼 보인 것이다. 칸반 Phase 1/2 Playwright 검증에서 `finalReportPath`가 비어 보이는 증상도 이 문제였다 ([kanban-phase1-2-playwright-validation 참조](./plans/2026-03-16-kanban-phase1-2-playwright-validation.md)).

해결은 단순했다. 포트 점유 PID 확인 -> 종료 -> 재시작 -> 시작 로그/헬스체크 확인 순서를 체크리스트화했다. 교훈은 "기능 버그를 의심하기 전에 런타임 환경부터 확인하라"는 것이다.

### Gateway 오류가 HTTP 200으로 숨겨지던 문제

`curl /api/sessions` 응답이 200 OK인데 body에 `{"error":"gateway rpc error..."}`가 들어 있었다. Gin 로그에도 200만 찍혀서 오류인지 정상인지 구분이 안 됐다. 원인은 Gateway 장애 시에도 프론트엔드 fallback을 위해 항상 200을 반환하도록 설계했기 때문이다.

502로 전환을 시도했지만, 프론트엔드의 파싱 구조와 맞지 않았다. 결국 200을 유지하되 `slog.Warn`으로 Gateway 상태와 에러 메시지를 출력하도록 타협했다 ([gateway-client.go 참조](./snippets/gateway-client.go.txt)). BFF 패턴에서 업스트림 오류를 투명하게 전달할지, fallback으로 감쌀지는 초기에 정해야 한다는 교훈을 얻었고, 이를 [CLAUDE.md](./CLAUDE.md)에 "200 OK fallback" 예외 사항으로 명시했다.

### 칸반 오케스트레이션 허브의 결합도

칸반 기능이 Phase 7까지 진행되면서, `completeRunSuccessLocked` 하나의 함수가 run 종료, artifact sync, review 판정, parent refresh, auto-finish를 모두 처리하게 됐다. 단일 함수의 blast radius가 너무 커서, 한 곳을 수정하면 연쇄적으로 영향이 퍼졌다.

이를 해결하기 위해 "behavior-preserving structural stabilization" 전략을 채택했다. 새 정책 도입 없이, 현재 의미를 유지한 채 허브 함수를 단계별로 분리하는 것이다. claim matrix로 사실과 가설을 구분하고, canonical semantics 보존 규칙을 정의한 뒤, Phase 1~4로 나눠서 진행했다 ([kanban-stabilization-plan 참조](./plans/2026-03-24-kanban-stabilization-plan.md)).

---

## 5. 결과 — "결국 뭐가 나왔나?"

### 아키텍처

```
┌───────────────────┐  REST / SSE   ┌──────────────┐  WebSocket   ┌──────────────────┐
│  Svelte 5 SPA     │ ───────────>  │  Go BFF      │  JSON-RPC    │  OpenClaw        │
│  (Vite + TS)      │ <───────────  │  (Gin :3001) │ <──────────> │  Gateway         │
│  :5173            │               └──────────────┘              └──────────────────┘
└───────────────────┘
     브라우저                          백엔드                        에이전트 런타임
```

| 레이어 | 기술 |
|--------|------|
| 프론트엔드 | Svelte 5 (Runes) + Vite + TypeScript |
| 백엔드 | Go + Gin |
| 실시간 통신 | WebSocket (BFF <-> Gateway), SSE (브라우저 <-> BFF) |

### 주요 기능

- **AI 에이전트 세션 실시간 모니터링**: SSE 기반 이벤트 스트리밍, keepalive로 프록시/브라우저 타임아웃 방지 ([sse-handler.go 참조](./snippets/sse-handler.go.txt))
- **Agent Board**: 칸반 기반 작업 오케스트레이션. parent/child 카드 관계 강조, guardrail 배지, 우선순위 표시 ([agent-board-dashboard.svelte 참조](./snippets/agent-board-dashboard.svelte))
- **채팅 인터페이스**: 서버 권위 큐 기반 메시지 전송, 인라인 메시지 렌더링
- **사용량 분석 대시보드**: 토큰 사용량 시각화 (SQLite + Chart.js)
- **크론 잡 관리**: Gateway 크론 작업 생성/조회/삭제

### 진행 규모

- MVP (Phase 1~5) + Kanban/Agent Board (Phase 1~7) 완료
- 플랜 문서 50개 이상 생성
- CSS 토큰 기반 다크/라이트 테마 시스템 구축
- Gateway WebSocket 클라이언트: challenge 핸드셰이크, exponential backoff 재연결, 동시성 안전 RPC ([gateway-client.go 참조](./snippets/gateway-client.go.txt))
