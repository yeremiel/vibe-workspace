# OpenClaw Workbench 운영 메모 (repo: claw-dash)

## 실행 커맨드

- `make install`: 백엔드/프론트 의존성 설치
- `make dev-backend`: 백엔드 단독 실행 (`:3001`)
- `make dev-frontend`: 프론트 단독 실행 (`:5173`)
- `make dev`: 백엔드+프론트 동시 실행
- `make verify-mvp`: Mock Gateway 포함 필수 검증 자동 실행
- `make security`: 정적 점검 (`go vet` + `npm audit --omit=dev`)
- `make ci` (`make qa`): 빌드/테스트/보안/MVP 검증 통합 실행

## 환경 변수

- `OPENCLAW_GATEWAY_URL`: Gateway WebSocket 주소
- `OPENCLAW_GATEWAY_TOKEN`: Gateway 토큰
- `OPENCLAW_DEVICE_IDENTITY_PATH`: 디바이스 키/ID 저장 경로 (기본 `.claw-dash-device.json`, 파일명은 호환성 유지)
- `OPENCLAW_RPC_TIMEOUT`: Gateway RPC 요청 타임아웃 (기본 `8s`)
- `BIND_HOST`: 백엔드 바인딩 호스트 (기본 `127.0.0.1`)
- `PORT`: 백엔드 HTTP 포트 (기본 `3001`)
- `OCL_AGENTS_DIR`: Usage 집계 대상 agents 디렉터리 (기본 `$HOME/.openclaw/agents`)
- `OCL_DB_PATH`: Usage 캐시 SQLite 파일 경로 (기본 `usage_cache.db`)

## 운영 원칙

- 로컬 단일 사용자 기준으로 최소 구성을 유지한다.
- 장애 확인은 `/health`와 `make verify-mvp`를 우선 사용한다.

## Knowledge Base 참조

공통 표준(`common/`)은 워크스페이스 CLAUDE.md에서 전체 적용. 이 프로젝트에서 추가로 참조하는 언어별 문서:

| 문서 | 경로 | 내용 |
|------|------|------|
| Go 코딩 컨벤션 | `@../../../knowledge_base/go/conventions.md` | 네이밍, 에러 처리, import 정렬, 테스트 작성 규칙 |
| Go 디자인 패턴 | `@../../../knowledge_base/go/design-patterns.md` | Clean Architecture 계층 분리, 인터페이스 추상화, 동시성 패턴 |

### BFF 특성상 예외 사항

이 프로젝트는 OpenClaw Gateway를 중계하는 BFF로, 아래 항목은 의도적으로 표준과 다르게 적용한다.

- **Response Envelope**: Gateway 응답을 그대로 중계하므로 `status/code/data/metadata` envelope 미적용
- **200 OK fallback**: Gateway 장애 시 빈 목록 + `fallback: true` + `error` 필드를 HTTP 200으로 반환한다. 프론트엔드가 Gateway 상태를 `gateway` / `fallback` 필드로 판별할 수 있도록 하기 위한 의도된 설계이며, `api-response.md` 원칙 4("2xx에 에러를 싣지 않는다")의 의도적 예외다.
- **Service 레이어**: Gateway RPC 호출만 수행하므로 Handler → Client 직접 호출 허용
- **Swagger 주석**: 내부 BFF로 프론트엔드와 1:1 통신, Swagger 문서화 생략

## 백엔드 구조 규칙

- `backend/cmd/server/main.go`: 설정 로드, signal 처리, 서버 lifecycle만 담당
- `backend/internal/app`: 공용 의존성 조립 및 시작/종료 관리
- `backend/internal/router/<domain>`: URL 등록만 담당
- `backend/internal/handler/<domain>`: HTTP 바인딩, 검증, 응답 작성
- `backend/internal/service/<domain>`: 도메인 로직과 orchestration
- `backend/internal/repository/<domain>`: DB/file 기반 영속화
- `backend/internal/model/<domain>`: 공용 DTO/모델

### 테스트 규칙

- 단위 테스트는 대상 패키지 옆 `*_test.go` 유지
- `backend/test/integration`: 진짜 라우터/앱 wiring 검증
- `backend/test/mocks`: 패키지 간 공유 mock/fake/fixture

## 프론트엔드 테마 규칙

### 기본 원칙

- 신규 컴포넌트 `<style>` 블록에 색상 하드코딩 금지 → 반드시 CSS 토큰 사용
- `--muted`는 deprecated → `--text-muted` 사용
- 전역 transition (`background-color 280ms`, `border-color 280ms`, `color 200ms`)이 `*` 선택자에 이미 적용되어 있으므로 컴포넌트에서 동일 속성 transition 재정의 불필요

### CSS 토큰 레퍼런스

| 용도 | 토큰 |
|------|------|
| 최외곽 배경 | `--bg-base` |
| 패널/카드 배경 | `--panel` |
| 패널 테두리 | `--panel-border` |
| 입력/셀 배경 | `--fill-base` |
| 살짝 띄운 표면 | `--fill-raised` |
| 텍스트 계층 | `--text-primary` / `--text-secondary` / `--text-muted` / `--text-faint` |
| 일반 테두리 | `--border-base` |
| 미묘한 테두리 | `--border-subtle` |
| 인터랙션 테두리 | `--border-interactive` |
| 버튼 배경 | `--btn-bg` (gradient) |
| 버튼 활성화 | `--btn-bg-active` |
| 선택 상태 | `--selected-border` / `--selected-bg` |
| 채팅 역할색 | `--role-user` / `--role-agent` |
| 강조 파랑 | `--accent-blue` |
| 강조 초록 | `--accent-green` |
| 초록 채널값(JS용) | `--accent-green-rgb` |
| 경고 주황 | `--accent-orange` |
| 위험/오류 | `--danger` |
| 네이티브 UI 테마 | `--color-scheme` |
| 히트맵 빈 셀 | `--cell-empty` |
| 파랑 펄스 애니메이션 | `--accent-blue-pulse` |

### 패턴 A: `<style>` 블록 — 일반 CSS

```css
/* 올바름 */
.panel { background: var(--panel); border: 1px solid var(--border-base); }
.label { color: var(--text-muted); }
.btn { background: var(--btn-bg); }
.btn.active { background: var(--btn-bg-active); border-color: var(--selected-border); }

/* 금지 */
.panel { background: rgba(14,24,34,0.9); }
```

### 패턴 B: JS에서 CSS 토큰 값 읽기

Chart.js, Canvas, 인라인 스타일 등 JS에서 색상이 필요할 때:

```ts
const cssVar = (name: string) =>
  getComputedStyle(document.documentElement).getPropertyValue(name).trim();

// 사용 예
cssVar('--accent-green')                        // '#4de88e' (다크) / '#1a9e5e' (라이트)
`rgba(${cssVar('--accent-green-rgb')}, 0.15)`   // rgba() 조합용
```

- `--accent-green-rgb`는 rgb 채널값만 담고 있음 (`77, 232, 142` 형태)
- `rgba()` 합성이 필요한 경우에만 별도 rgb 토큰을 사용한다

### 패턴 C: Chart.js — 테마 전환 동기화

Chart.js는 생성 시점에 색상을 고정하므로, `data-theme` 변경을 감지해 수동 업데이트해야 한다:

```ts
import { onMount } from 'svelte';

// 차트 생성 시: 즉시 읽어 주입
datasets: [{ borderColor: cssVar('--accent-green') }]

// 테마 전환 감지: MutationObserver
onMount(() => {
  const obs = new MutationObserver(() => {
    if (!chart) return;
    (chart.data.datasets[0] as Record<string, unknown>)['borderColor'] = cssVar('--accent-green');
    chart.update('none'); // 'none' = 애니메이션 없이 즉시 갱신
  });
  obs.observe(document.documentElement, { attributes: true, attributeFilter: ['data-theme'] });
  return () => obs.disconnect();
});
```

글로벌 Chart.js 기본값은 `chartjs-setup.ts`의 `applyChartTheme()`이 처리하므로 grid/label 색상은 컴포넌트에서 별도 처리 불필요

### 파생 색상 합성

토큰에 없는 색상이 필요한 경우 `color-mix()` 사용:

```css
/* 오류 배경: fill-base와 danger 혼합 */
.toast { background: color-mix(in srgb, var(--fill-base) 90%, var(--danger) 10%); }

/* 반투명 오버레이: rgba(하드코딩) 대신 bg-base 기반으로 테마에 따라 자동 조정 */
.backdrop { background: color-mix(in srgb, var(--bg-base) 42%, transparent); }
```

### 네이티브 UI 요소

`<input type="date">`, 스크롤바, `<select>` 등 OS 네이티브 UI는 `color-scheme` 속성을 사용:

```css
.date-input { color-scheme: var(--color-scheme); } /* 다크: dark, 라이트: light */
```

`html { color-scheme: var(--color-scheme); }`는 `app.css`에 전역으로 이미 설정되어 있음.

### 애니메이션 (keyframes)

keyframe 내부에서도 CSS 토큰 사용:

```css
@keyframes pulse {
  0%, 100% { box-shadow: 0 0 0 0 var(--accent-blue-pulse); }
  50%       { box-shadow: 0 0 0 8px transparent; }
}
```
