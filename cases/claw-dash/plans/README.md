# 플랜 문서 목록

프로젝트 전체에서 생성된 50개 이상의 플랜 문서 중 대표적인 것을 유형별로 공개한다.

## 아키텍처 설계

| 문서 | 내용 |
|------|------|
| [backend-structure-refactor-design](./2026-03-20-backend-structure-refactor-design.md) | Handler → Service → Repository 계층 분리 리팩터링 설계 |

## 기능 구현

| 문서 | 내용 |
|------|------|
| [session-chat-inline-message-design](./2026-03-19-session-chat-inline-message-design.md) | 채팅 인라인 메시지 렌더링 설계 |

## 안정화 / 좌충우돌

| 문서 | 내용 |
|------|------|
| [kanban-stabilization-plan](./2026-03-24-kanban-stabilization-plan.md) | 칸반 오케스트레이션 허브 안정화 (claim matrix, phase별 scope lock) |

## QA / 검증

| 문서 | 내용 |
|------|------|
| [kanban-phase1-2-playwright-validation](./2026-03-16-kanban-phase1-2-playwright-validation.md) | Playwright 기반 칸반 Phase 1/2 smoke 검증 |

## 칸반 시스템 (별도 폴더)

칸반이 이 프로젝트에서 가장 복잡한 서브시스템이었다.
설계 → 진화 → 리팩터링 → 자율화 시도까지의 흐름을 담은 문서들을 [`kanban/`](./kanban/)에 별도로 모았다.

| 문서 | 내용 |
|------|------|
| [kanban-system](./kanban/kanban-system.md) | 칸반 상태 머신 및 전체 아키텍처 |
| [kanban-artifact-lifecycle](./kanban/kanban-artifact-lifecycle.md) | 아티팩트 라이프사이클 상세 규칙 |
| [kanban-waterfall-flow-realignment](./kanban/2026-03-14-kanban-waterfall-flow-realignment.md) | 에이전트 드리븐 칸반 흐름 설계 |
| [kanban-refactoring-phase0-6-summary](./kanban/2026-03-17-kanban-refactoring-phase0-6-summary.md) | 13K LOC 리팩터링 Phase 0~6 요약 |
| [autonomous-kanban-v2-evidence-contract](./kanban/2026-03-24-autonomous-kanban-v2-evidence-contract.md) | 자율 완료를 위한 executor evidence contract 설계 |
| [autonomous-kanban-v2-evidence-gap-analysis](./kanban/2026-03-24-autonomous-kanban-v2-evidence-gap-analysis.md) | 기존 구현의 gap 분석 |
