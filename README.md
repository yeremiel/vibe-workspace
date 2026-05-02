# vibe-workspace

사람과 AI 에이전트(Claude Code)가 같은 문서를 읽고 같은 규칙으로 일하는, **살아있는 개발 표준 워크스페이스**입니다.

[🇯🇵 日本語](README.ja.md) | 🇰🇷 한국어

## 이 저장소는 무엇인가요?

프로젝트 모음이 아니라 **워크스페이스 자체가 산출물**입니다.

- 프로젝트가 늘어날수록 표준과 가이드라인에 살이 붙는다
- 사람도 AI 에이전트도 **같은 문서**를 참조해 같은 결과물을 만든다
- 새 프로젝트를 시작해도 컨텍스트를 다시 설명할 필요가 없다

이 저장소가 보여주는 것은 "어떤 프로젝트를 만들었나"가 아니라, **그 프로젝트들이 자라는 토양을 어떻게 설계했나**입니다.

## 워크스페이스의 핵심: 살아있는 개발 표준

`knowledge_base/`에는 언어/도메인을 가로지르는 개발 표준 문서가 누적되어 있습니다. 어떤 프로젝트를 시작하든, AI 에이전트는 이 문서들을 먼저 읽고 코드를 생성합니다.

| 문서 | 적용 범위 |
|------|-----------|
| [`common/api-response.md`](knowledge_base/common/api-response.md) | 모든 백엔드 API의 응답 envelope·에러 코드 체계 — 언어/프레임워크 무관 동일 적용 |
| [`common/git-workflow.md`](knowledge_base/common/git-workflow.md) | 모든 프로젝트 공통 브랜치 전략·커밋 메시지 컨벤션 |
| [`go/conventions.md`](knowledge_base/go/conventions.md) | Go 백엔드 작성 시 네이밍·에러 처리·로깅 규칙 |
| [`go/design-patterns.md`](knowledge_base/go/design-patterns.md) | Go 프로젝트 아키텍처 패턴 (Clean Architecture, DI 등) |
| [`flutter/conventions.md`](knowledge_base/flutter/conventions.md) | Flutter 앱 작성 시 위젯·상태 관리·라우팅 규칙 |
| [`flutter/design-patterns.md`](knowledge_base/flutter/design-patterns.md) | Flutter Feature-First 구조와 Riverpod 패턴 |
| [`flutter/gotchas.md`](knowledge_base/flutter/gotchas.md) | Flutter 작업 중 반복적으로 마주친 함정과 해결 방식 |

### 어떻게 적용되는가

새 백엔드를 시작할 때:

1. 프로젝트 루트의 `CLAUDE.md`에 **"이 프로젝트는 `knowledge_base/go/`와 `knowledge_base/common/`을 따른다"** 한 줄
2. 그 순간부터 AI는 모든 엔드포인트를 동일한 응답 envelope로, 동일한 에러 코드 체계로 작성
3. 다른 백엔드 프로젝트로 옮겨가도 같은 문서를 참조 → **같은 송수신 스펙이 자동으로 강제됨**

### 살아있다는 것

표준은 한 번 만들고 끝이 아닙니다. 실제 프로젝트에서 발견한 패턴/실수가 다시 표준에 반영되는 피드백 루프를 가집니다.

- Flutter `gotchas.md`는 작업 중 만난 함정을 누적한 결과물
- Go `design-patterns.md`는 여러 백엔드를 거치며 정제된 구조 결정
- 새로운 언어/도메인이 추가되면 `knowledge_base/` 아래 새 폴더가 늘어남

## 컨텍스트 계층 구조: CLAUDE.md

`CLAUDE.md`는 사람과 AI가 함께 읽는 **컨텍스트 진입점**입니다. 워크스페이스 → 프로젝트로 내려가면서 점점 구체화되는 계층 구조를 가집니다.

```
워크스페이스 CLAUDE.md       ← 작업자/환경/공통 금지 사항
    └── 프로젝트 CLAUDE.md     ← 해당 프로젝트의 스택·아키텍처·테스트 규칙
            └── knowledge_base/ 참조  ← 언어·도메인별 표준 위임
```

상위는 변하지 않고, 하위는 프로젝트마다 다르며, 표준은 한 곳에 모여 있습니다. 사람이 새 멤버에게 온보딩할 때 보여주는 문서와, AI 에이전트가 작업 시작 시 읽는 문서가 완전히 동일합니다.

설계 사상의 상세 → [environment/workspace-structure.md](environment/workspace-structure.md)

## 환경 구조

| 디렉터리 | 역할 |
|----------|------|
| `environment/` | 워크스페이스 AI 컨텍스트 + 구조 설계 철학 |
| `knowledge_base/` | 언어별·도메인별 개발 표준 (위 참조) |
| `cases/` | 표준 시스템이 실제 프로젝트에서 어떻게 굴러갔나의 사례 |

## 사례: 시스템이 실제로 작동하는가

### [OpenClaw Workbench](cases/claw-dash/)

AI 에이전트 오케스트레이션 대시보드. Go BFF + Svelte 5 SPA 구조로,
OpenClaw Gateway와 WebSocket으로 연결해 실시간 에이전트 세션을 시각화한다.

이 케이스는 위에서 설명한 워크스페이스 구조가 실제 프로젝트 진행 중 어떻게 작동했는지를 추적합니다 — 컨텍스트를 어떻게 구성했고, AI와 어디서 헤맸으며, 어떤 규칙을 만들어 풀었는지. 프로젝트 자체의 자랑이 아니라 **표준 시스템의 검증 사례**입니다.

## 권장 읽기 순서

1. **[지금 여기]** 워크스페이스 개요 (이 파일)
2. **[핵심]** [knowledge_base/](knowledge_base/) — 실제 표준 문서들
3. **[보조]** [environment/CLAUDE.md](environment/CLAUDE.md) — 컨텍스트 계층 원본
4. **[사례]** [cases/claw-dash/](cases/claw-dash/) — 시스템 검증 케이스 스터디
