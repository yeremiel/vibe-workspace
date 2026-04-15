# vibe-workspace

AI 에이전트(Claude Code)와 함께 일하는 방식을 정리한 개인 개발 환경 포트폴리오입니다.

## 이 저장소는 무엇인가요?

단순한 프로젝트 모음이 아닙니다. AI 에이전트와 협업하기 위해 **환경 자체를 어떻게 설계하고 운영하는지**를 보여줍니다.

- AI에게 어떤 컨텍스트를 주고 어떤 규칙으로 일했는지
- 기능 단위로 어떻게 계획하고 AI와 함께 실행했는지
- 실제 프로젝트에서 어떤 문제를 만나고 어떻게 헤쳐나갔는지

## 권장 읽기 순서

1. **[지금 여기]** 환경 설계 개요 (이 파일)
2. **[핵심]** [cases/claw-dash/](cases/claw-dash/) — 실제 프로젝트 케이스 스터디
3. **[참고]** [environment/CLAUDE.md](environment/CLAUDE.md) — AI 컨텍스트 설계 원본
4. **[참고]** [knowledge_base/](knowledge_base/) — 개발 표준

## 환경 구조

| 디렉터리 | 역할 |
|----------|------|
| `environment/` | 워크스페이스 AI 컨텍스트 + 구조 설계 철학 |
| `knowledge_base/` | 언어별·도메인별 개발 표준 |
| `cases/` | 프로젝트별 케이스 스터디 |

자세한 구조 설계 이유 → [environment/workspace-structure.md](environment/workspace-structure.md)

## 케이스 스터디

### [OpenClaw Workbench](cases/claw-dash/)

AI 에이전트 오케스트레이션 대시보드. Go BFF + Svelte 5 SPA 구조로,
OpenClaw Gateway와 WebSocket으로 연결해 실시간 에이전트 세션을 시각화한다.

이 프로젝트에서 AI와 함께 가장 많이 헤맸고, 가장 많이 배웠다.

## 병행 중인 프로젝트

- **myLib** — 개인 도서 관리 서비스 (Go + Flutter, 비공개)
- **scanner** — Flutter + OpenCV 오프라인 문서 스캐너 (비공개)
