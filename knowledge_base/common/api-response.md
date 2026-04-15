# API Response & Error Standard (v1)

이 문서는 **백엔드 API의 응답 포맷(Response Envelope)** 과 **에러 코드 체계(Error Code System)** 를 하나의 표준으로 정의한다.  
언어/프레임워크(Go, Java, Node 등)와 무관하게 **모든 백엔드 API**에서 동일하게 적용한다.

> 목표: 클라이언트(Flutter/Web/SDK)가 `status / code / error / message`만으로 **일관된 분기 처리**를 할 수 있도록 한다.

---

## 1. 핵심 원칙 (Principles)

1. **HTTP Status = code**  
   - 응답 본문의 `code`는 **항상 HTTP status와 동일**해야 한다.
2. **Machine-readable vs Human-readable 분리**  
   - `error`: 기계가 분기하기 위한 **표준 코드(standardized code)**  
   - `message`: 사용자 표시/로깅용 **메시지(i18n key 또는 plain message)**
3. **Backwards compatible를 기본으로**  
   - 필드 확장은 가능하되, 의미 변경/파괴적 변경(breaking change)은 최소화한다.
4. **2xx에 에러를 싣지 않는다**  
   - `status="error"`인데 HTTP 200인 응답은 금지.

---

## 2. Response Envelope (응답 구조)

### 2.1 Success Response (성공)

```json
{
  "status": "success",
  "code": 200,
  "message": "ok",
  "data": { },
  "metadata": {
    "timestamp": "2026-02-12T10:30:00Z",
    "traceId": "abc123"
  }
}
```

### 2.2 Error Response (실패)

```json
{
  "status": "error",
  "code": 401,
  "error": "AUTH_TOKEN_EXPIRED",
  "message": "auth.token_expired",
  "metadata": {
    "timestamp": "2026-02-12T10:30:00Z",
    "traceId": "abc123"
  }
}
```

---

## 3. 필드 정의 (Field Definitions)

| 필드 | 타입 | 필수 | 목적 | 규칙 |
|---|---|---:|---|---|
| status | string | O | 성공/실패 구분 | `"success"` 또는 `"error"` |
| code | number | O | HTTP status mirror | **항상 HTTP status와 동일** |
| data | object/array/primitive | △ | 성공 데이터 | **성공에서만 사용**. 데이터가 없으면 생략(omit) 권장 |
| error | string | △ | 분기용 에러 코드 | **실패에서 권장**. 표준 목록 준수 |
| message | string | O | 표시/설명 메시지 | i18n 사용 시 key, 미사용 시 plain message 허용 |
| errors | array | △ | 필드 단위 검증 에러 | **422에서 권장** (아래 7장 참고) |
| metadata | object | O | 공통 메타 | timestamp 필수, traceId/requestId 등 확장 가능 |

### 3.1 `message` 규칙 (i18n / non-i18n)

- i18n을 사용하는 서비스:
  - `message`는 **i18n key** (예: `auth.token_expired`)
- i18n을 사용하지 않는 서비스:
  - `message`는 **plain human message** 허용 (예: `Token expired`)

> 클라이언트 표시 전략은 프로젝트 성격에 따라 다르므로, `message`는 **표시 가능한 값**을 보장한다.

### 3.2 `data` 규칙 (omit vs null)

- 반환 데이터가 없으면 `data`는 **생략(omit) 권장**
- `null`은 가능하지만, 클라이언트 분기 처리가 복잡해지므로 권장하지 않는다.

---

## 4. Metadata 표준 (Observability)

`metadata`는 관측/추적(Observability)을 위해 사용한다.

### 4.1 필수 필드

- `metadata.timestamp`: RFC3339/ISO8601 (예: `2026-02-12T10:30:00Z`)

### 4.2 권장 확장

- `metadata.traceId`: 분산 트레이싱/요청 추적
- `metadata.requestId`: 요청 식별자(게이트웨이/서버에서 생성)
- `metadata.version`: 응답 스키마/서버 버전(옵션)

예:

```json
"metadata": {
  "timestamp": "2026-02-12T10:30:00Z",
  "traceId": "abc123",
  "requestId": "req-9f8e",
  "version": "v1"
}
```

---

## 5. Error Code System (에러 코드 체계)

### 5.1 네이밍 규칙 (Naming)

- 형식: `UPPER_SNAKE_CASE`
- 가능한 한 **구체적으로** (원인 + 대상 + 상황)
- 도메인/관심사 prefix 권장:

| Prefix | 의미 |
|---|---|
| REQUEST_* | 요청 형식/파라미터/검증 |
| AUTH_* | 인증/인가(Authentication/Authorization) |
| INTERNAL_* | 서버 내부/의존성 장애 |
| <DOMAIN>_* | 도메인별(User/Book/Order 등) |

예:
- `REQUEST_MALFORMED_JSON`
- `AUTH_INVALID_TOKEN`
- `USER_EMAIL_ALREADY_EXISTS`
- `BOOK_NOT_FOUND`

### 5.2 HTTP Status ↔ Error Code 매핑 규칙

> HTTP status는 “실패의 종류”, error는 “실패의 정확한 의미”를 나타낸다.

| HTTP | 의미 | 대표 error 코드 예시 |
|---|---|---|
| 400 Bad Request | 잘못된 요청(형식/파싱) | `REQUEST_INVALID`, `REQUEST_MALFORMED_JSON` |
| 401 Unauthorized | 인증 실패 | `AUTH_UNAUTHORIZED`, `AUTH_INVALID_TOKEN`, `AUTH_TOKEN_EXPIRED` |
| 403 Forbidden | 권한 없음 | `AUTH_FORBIDDEN`, `AUTH_SCOPE_INSUFFICIENT` |
| 404 Not Found | 리소스 없음 | `RESOURCE_NOT_FOUND`, `USER_NOT_FOUND`, `BOOK_NOT_FOUND` |
| 409 Conflict | 충돌/중복 | `RESOURCE_CONFLICT`, `USER_EMAIL_ALREADY_EXISTS` |
| 422 Unprocessable Entity | 의미/도메인 검증 실패 | `REQUEST_VALIDATION_FAILED`, `BOOK_INVALID_ISBN` |
| 429 Too Many Requests | Rate limit | `REQUEST_RATE_LIMITED` |
| 500 Internal Server Error | 서버 내부 오류 | `INTERNAL_ERROR` |
| 503 Service Unavailable | 의존성 장애/점검 | `INTERNAL_DEPENDENCY_UNAVAILABLE` |

### 5.3 금지/주의 규칙

- `status="error"`인데 `code=200` 금지
- 4xx는 클라이언트 수정/재시도로 해결 가능한 오류, 5xx는 서버/인프라 원인으로 구분
- 같은 조건에서는 **항상 같은 `error`**를 반환한다 (일관성)

---

## 6. 표준 Error Code 목록 (v1 최소 세트)

> 처음부터 모든 도메인을 완벽하게 만들기보다, **공통 + 인증/인가**부터 시작한다.

### 6.1 Common

- `REQUEST_INVALID` (400)
- `REQUEST_MALFORMED_JSON` (400)
- `REQUEST_RATE_LIMITED` (429)
- `INTERNAL_ERROR` (500)
- `INTERNAL_DEPENDENCY_UNAVAILABLE` (503)

### 6.2 Validation (422)

- `REQUEST_VALIDATION_FAILED` (422)

### 6.3 Auth (401/403)

- `AUTH_UNAUTHORIZED` (401) — 인증 필요
- `AUTH_INVALID_CREDENTIALS` (401) — 로그인 실패(자격 증명 불일치)
- `AUTH_INVALID_TOKEN` (401) — 토큰 형식/서명 불일치
- `AUTH_TOKEN_EXPIRED` (401) — 토큰 만료
- `AUTH_FORBIDDEN` (403) — 권한 없음
- `AUTH_SCOPE_INSUFFICIENT` (403) — scope/role 부족

> 도메인별(`USER_*`, `BOOK_*` 등)은 실제 필요가 생기는 시점에 추가한다.

---

## 7. Validation Errors (422) 상세 규격

422에서 사용자 입력 검증 결과를 UI에 정확히 매핑할 수 있도록 `errors[]` 확장을 권장한다.

```json
{
  "status": "error",
  "code": 422,
  "error": "REQUEST_VALIDATION_FAILED",
  "message": "request.validation_failed",
  "errors": [
    { "field": "email", "reason": "invalid_format", "message": "user.email.invalid" },
    { "field": "password", "reason": "too_short", "message": "auth.password.too_short" }
  ],
  "metadata": { "timestamp": "2026-02-12T10:30:00Z", "traceId": "abc123" }
}
```

규칙:
- `errors[]`는 **422에서만 사용(권장)**
- `field`는 request payload key와 동일하게 유지
- `reason`은 UI 분기용(짧은 토큰)으로 유지 (`invalid_format`, `too_short` 등)
- `message`는 i18n key 권장, 미사용 시 plain message 허용

---

## 8. Client Handling Rule (클라이언트 처리 규칙)

클라이언트는 아래 우선순위로 처리하는 것을 권장한다.

1. `status`로 성공/실패 분기
2. 실패 시:
   - `error`가 있으면 **에러 코드 기반 분기** (예: 로그인 화면 이동, 재시도 등)
   - `message`는 사용자 표시/로그에 사용
3. `error`가 없으면:
   - `message`로 fallback (호환성 유지)

---

## 9. Implementation Guideline (서버 구현 가이드)

### 9.1 Single Source of Truth

- 에러 코드는 enum/const로 관리한다.
- 핸들러/컨트롤러는 예외/에러를 받아 **(httpStatus, errorCode, message)** 로 변환하는 Mapper를 사용한다.

### 9.2 Logging

- 에러 로그에는 최소 다음을 포함한다.
  - `code`, `error`, `traceId`(또는 requestId), 원인 스택/컨텍스트

---

## 10. Change Policy (변경 정책)

- 이 문서의 변경은 API Contract 변경이므로, 변경 시 반드시:
  - 서버/클라이언트 영향 범위 확인
  - 버전 태그/마이그레이션 전략 기록(필요 시)
  - 테스트(서버/클라이언트) 업데이트

---

## Appendix A. Examples (예시)

### A.1 Not Found

```json
{
  "status": "error",
  "code": 404,
  "error": "BOOK_NOT_FOUND",
  "message": "book.not_found",
  "metadata": { "timestamp": "2026-02-12T10:30:00Z", "traceId": "abc123" }
}
```

### A.2 Conflict

```json
{
  "status": "error",
  "code": 409,
  "error": "USER_EMAIL_ALREADY_EXISTS",
  "message": "user.email_already_exists",
  "metadata": { "timestamp": "2026-02-12T10:30:00Z", "traceId": "abc123" }
}
```
