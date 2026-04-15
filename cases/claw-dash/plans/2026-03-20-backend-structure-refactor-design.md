# Backend Structure Refactor Design

## 배경

`claw-dash` 백엔드는 기능이 빠르게 추가되면서 `cmd/server/main.go`에 의존성 조립과 라우팅이 집중되었고, HTTP 핸들러도 `internal/handler` 단일 디렉토리에 누적되었다. 반면 실제 비즈니스 로직은 `internal/kanban`, `internal/usage`, `internal/gateway`로 흩어져 있어, 책임 경계와 파일 배치 규칙이 일관되지 않다.

이번 리팩터링의 목표는 `myLib-backend`와 유사한 책임 분리를 도입하되, Go 테스트 관례와 현재 코드베이스의 내부 상태 검증 요구를 반영해 구조를 재정렬하는 것이다.

## 목표

- `cmd/server/main.go`를 얇게 유지하고 앱 조립 책임을 `internal/app`으로 이동한다.
- URL 등록 책임을 `internal/router`로 분리한다.
- HTTP 엔드포인트를 도메인별 `internal/handler/<domain>` 패키지로 분리한다.
- 핵심 도메인 로직을 `internal/service/<domain>` 패키지로 재배치한다.
- 테스트 구조를 정리하되, Go 특성상 단위 테스트는 대상 패키지 옆에 유지한다.
- 기존 API 경로와 JSON 응답 shape는 유지한다.

## 목표 구조

```text
backend/
├── cmd/
│   ├── mock-gateway/
│   └── server/
├── internal/
│   ├── app/
│   ├── config/
│   ├── gateway/
│   ├── handler/
│   │   ├── agents/
│   │   ├── cron/
│   │   ├── events/
│   │   ├── health/
│   │   ├── kanban/
│   │   ├── nodes/
│   │   ├── ops/
│   │   ├── sessions/
│   │   ├── skills/
│   │   ├── uploads/
│   │   ├── usage/
│   │   └── voice/
│   ├── router/
│   │   ├── agents/
│   │   ├── cron/
│   │   ├── events/
│   │   ├── health/
│   │   ├── kanban/
│   │   ├── nodes/
│   │   ├── ops/
│   │   ├── sessions/
│   │   ├── skills/
│   │   ├── uploads/
│   │   ├── usage/
│   │   └── voice/
│   ├── service/
│   │   ├── gatewayrpc/
│   │   ├── kanban/
│   │   ├── sessions/
│   │   └── usage/
│   └── model/
│       └── shared/
└── test/
    ├── integration/
    └── mocks/
```

## 책임 분리 원칙

### internal/app

- 설정 로드 이후 공용 의존성 초기화
- gateway client, usage DB, kanban store, project registry, session send queue 조립
- 앱 시작/종료 lifecycle 관리

### internal/router

- Gin route 등록만 담당
- 도메인별 `Register` 함수로 `/health`, `/api/...` 경로를 연결
- 비즈니스 로직 없음

### internal/handler/<domain>

- HTTP 요청 바인딩과 검증
- HTTP 상태 코드 및 JSON 응답 작성
- 서비스 호출 결과를 API 응답으로 변환

### internal/service/<domain>

- 도메인 로직 및 워크플로우 유지
- `usage`, `kanban`의 기존 로직을 서비스 계층으로 재배치
- gateway RPC retry / decode / metrics는 `service/gatewayrpc` 공용 패키지로 분리

### internal/model

- 공용 API 응답 메타 등 재사용 DTO를 점진적으로 이동
- 이번 리팩터링에서는 최소 공용 타입부터 도입

## 도메인 매핑

- `handler/health.go` -> `handler/health`
- `handler/events.go` -> `handler/events`
- `handler/uploads.go` -> `handler/uploads`
- `handler/voice.go` -> `handler/voice`
- `handler/ops_health.go` -> `handler/ops`
- `handler/nodes_status.go` -> `handler/nodes`
- `handler/cron.go` -> `handler/cron`
- `handler/sessions.go`, `session_history.go`, `session_admin.go` -> `handler/sessions`
- `handler/session_send_queue.go` -> `service/sessions`
- `handler/skills_*.go` -> `handler/skills`
- `handler/agents_admin.go` -> `handler/agents`
- `internal/usage/*.go` -> `service/usage`
- `internal/kanban/*.go` -> `service/kanban`

## 테스트 원칙

- 단위 테스트는 각 패키지 옆 `*_test.go`로 유지한다.
- `backend/test/integration`은 HTTP 라우터, 앱 조립, 파일/DB 연동 같은 진짜 통합 테스트만 둔다.
- `backend/test/mocks`는 패키지 간 공유 mock/fake/fixture builder를 둔다.

이 방식은 `myLib-backend`의 구조적 의도를 유지하면서도, Go의 package-private 테스트 이점을 훼손하지 않는다.

## 구현 순서

1. `internal/app`, `internal/router`, 공용 `service/gatewayrpc` 도입
2. `main.go`를 앱 조립 진입점만 남도록 단순화
3. 결합도가 낮은 도메인부터 `handler/<domain>`, `router/<domain>`로 이동
4. `internal/usage`를 `service/usage`로 이동
5. `internal/kanban`을 `service/kanban`으로 이동
6. 라우터/앱 조립 갱신 후 `go test ./...`로 회귀 검증

## 비목표

- API 경로 변경
- 응답 JSON 필드 변경
- 기능 추가
- kanban 도메인의 내부 알고리즘 수정
