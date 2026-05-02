# knowledge_base — 살아있는 개발 표준

[🇯🇵 日本語](README.ja.md) | 🇰🇷 한국어

사람과 AI 에이전트(Claude Code)가 같은 문서를 읽고 같은 규칙으로 일하기 위한 개발 표준 모음입니다.

워크스페이스의 thesis 전체 → [최상위 README](../README.md)

## 어떻게 동작하는가

각 프로젝트의 `CLAUDE.md`에서 `@` 참조로 필요한 표준 문서만 가져옵니다.

```markdown
# 프로젝트 CLAUDE.md 예시

이 프로젝트는 다음 표준을 따른다.

@../../knowledge_base/common/api-response.md
@../../knowledge_base/common/git-workflow.md
@../../knowledge_base/go/conventions.md
@../../knowledge_base/go/design-patterns.md
```

이 한 번의 선언으로:

- AI 에이전트는 모든 코드 생성 시 해당 표준을 따름
- 사람(나, 또는 협업자)도 동일 문서로 온보딩
- 프로젝트가 바뀌어도 같은 송수신 스펙·같은 컨벤션 유지

## 디렉터리 구성

### `common/` — 언어 무관 표준

모든 프로젝트에 공통 적용. 백엔드든 프론트엔드든 동일.

| 문서 | 내용 |
|------|------|
| [`api-response.md`](common/api-response.md) | 백엔드 API 응답 envelope, 에러 코드 체계, HTTP status 매핑. **언어/프레임워크 무관 동일 적용** |
| [`git-workflow.md`](common/git-workflow.md) | 브랜치 전략, 커밋 메시지 컨벤션, 커밋 전 승인 절차 |

### `go/` — Go 백엔드 표준

| 문서 | 내용 |
|------|------|
| [`conventions.md`](go/conventions.md) | 패키지·함수·변수 네이밍, 에러 처리, 로깅, 테스트 작성 규칙 |
| [`design-patterns.md`](go/design-patterns.md) | 인터페이스 활용, Clean Architecture 계층 분리, DI, 동시성 패턴 |

### `flutter/` — Flutter/Dart 표준

| 문서 | 내용 |
|------|------|
| [`conventions.md`](flutter/conventions.md) | 파일·위젯·상태 네이밍, 라우팅, 디렉터리 구조 |
| [`design-patterns.md`](flutter/design-patterns.md) | Riverpod 3.x Notifier 패턴, Feature-First 구조, Freezed 활용 |
| [`gotchas.md`](flutter/gotchas.md) | 작업 중 마주친 함정과 해결 방식의 누적 (`copyWith` nullable 처리 등) |

## 살아있다는 것

표준은 **한 번 만들고 끝나는 정적 문서가 아닙니다.**

- 새 프로젝트에서 발견한 패턴이 표준으로 승격됨
- 반복적으로 마주친 실수가 `gotchas.md` 같은 문서로 누적됨
- 새 언어/도메인이 등장하면 이 디렉터리 아래에 새 폴더가 생김 (예: 향후 `java/`, `typescript/`)

이 진화 과정 자체가 워크스페이스의 산출물입니다. claw-dash 케이스 스터디에서는 BFF 특성으로 `api-response.md`의 일부 원칙을 의도적으로 예외 처리한 사례를 다룹니다 → [cases/claw-dash/](../cases/claw-dash/)

## 어디서부터 읽으면 좋은가

처음 방문이라면 다음 순서를 추천합니다.

1. **[`common/api-response.md`](common/api-response.md)** — 가장 강제력 있는 표준. "내가 만드는 모든 백엔드 API는 이렇게 응답한다"는 약속의 실체
2. **[`common/git-workflow.md`](common/git-workflow.md)** — 짧고 직관적. 작업 흐름의 전제
3. 본인 관심 스택의 `conventions.md` → `design-patterns.md` 순
4. (Flutter라면) [`flutter/gotchas.md`](flutter/gotchas.md) — 살아있는 시스템의 가장 직접적인 증거
