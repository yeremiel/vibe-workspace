# 워크스페이스 구조 설계

[🇯🇵 日本語](workspace-structure.ja.md) | 🇰🇷 한국어

AI 에이전트(Claude Code)와 함께 일하기 위해 설계한 로컬 워크스페이스 구조다.

## 디렉터리 구조

```
~/workspace/
├── knowledge_base/    # 프로젝트 횡단 개발 표준
├── obsidian/          # 작업 로그, 의사결정 기록, 리뷰 (Obsidian Vault)
├── projects/
│   └── self/          # 개인 토이프로젝트 모음
└── CLAUDE.md          # 워크스페이스 루트 AI 컨텍스트
```

## 설계 원칙

### knowledge_base/ — "AI에게 주는 공통 규칙"

언어별·도메인별 개발 표준을 한 곳에 모아 관리한다.
모든 프로젝트의 CLAUDE.md에서 `@` 참조로 불러온다.

- `common/`: API 응답 표준, Git 워크플로우 등 언어 무관 규칙
- `go/`: Go 코딩 컨벤션, 디자인 패턴
- `flutter/`: Flutter/Dart 컨벤션

이 구조 덕분에 새 프로젝트를 시작할 때 CLAUDE.md에서 표준 문서를 참조하기만 하면,
AI 에이전트가 일관된 코딩 스타일로 작업한다.

### obsidian/ — "작업 기록 허브"

프로젝트 진행 중 의사결정, 삽질 기록, Phase별 리뷰를 Obsidian 볼트로 관리한다.
코드 저장소에 들어가기 애매한 "왜 이렇게 했는가"를 남기는 공간이다.

### projects/self/ — "실험실"

외부 요구 없이 직접 기획하고 만드는 토이프로젝트 모음.
각 프로젝트마다 CLAUDE.md를 두어 프로젝트 특화 AI 컨텍스트를 관리한다.

## CLAUDE.md 계층 구조

```
workspace/CLAUDE.md              # 워크스페이스 공통 컨텍스트
└── projects/self/<project>/CLAUDE.md   # 프로젝트 특화 컨텍스트
    └── @knowledge_base/go/conventions.md  # 표준 문서 참조
```

AI 에이전트는 현재 작업 디렉터리의 CLAUDE.md를 읽고,
그 안에서 참조하는 knowledge_base 문서까지 함께 로드한다.
