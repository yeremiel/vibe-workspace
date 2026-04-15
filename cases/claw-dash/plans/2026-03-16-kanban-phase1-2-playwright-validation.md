# Kanban Phase 1/2 Playwright Validation (2026-03-16)

## 목적

Phase 1 (`implementation parent final-report`) 및 Phase 2 (`documentation parent finalized artifact chaining`)의 실전 검증 결과를 기록한다.

## 검증 환경

- Frontend: Vite dev server (`http://127.0.0.1:5173`)
- Backend: local dev server (`http://127.0.0.1:3001`)
- 검증 방식:
  - 테스트 task 생성: backend API
  - 실제 UI 확인: Playwright/browser snapshot

## 검증 시나리오

### 1. Documentation parent finalized chaining
- 테스트용 documentation parent 생성
- `planning -> review -> done` 전환
- `docs/task-<id>/final-proposal.md` 작성
- parent detail에서 finalized source 및 chaining entry 확인

### 2. Implementation parent final-report
- 테스트용 implementation parent 생성
- `planning -> review` 전환
- `docs/task-<id>/final-report.md` 자동 초안 생성 여부 확인
- parent detail에서 final report 대표 노출 여부 확인

## 결과

### Documentation parent
성공.

확인 내용:
- parent detail에서 `Final proposal` 패널 표시
- 경로 예시: `docs/task-364/final-proposal.md`
- `Create Intake from Finalized Artifact` 버튼 노출
- parent-only chaining 흐름이 실제 UI에서 동작 가능한 상태로 보임

판정:
- 문서형 parent finalized source 표시 및 chaining entry는 실전 기준으로 정상 동작

### Implementation parent
최종적으로 성공.

초기 관찰:
- 첫 검증 시 `finalReportPath` / `finalReport` 가 비어 보였음
- 이벤트에도 `final_report_draft_generated` 가 보이지 않았음

원인:
- 코드 결함이 아니라, backend가 최신 코드가 아닌 stale binary(go-build cache 실행본)로 떠 있었음
- 즉, 런타임 반영 문제였음

조치:
- stale backend 프로세스 종료
- 최신 코드 기준으로 backend 재시작
- 새 implementation parent(`task-367`)로 동일 시나리오 재검증

재검증 결과:
- `planning -> review` 전환 후 `finalReportPath = docs/task-367/final-report.md`
- detail API에서 `finalReport` 반환 확인
- `final-report.md` 자동 초안 내용 확인
  - 작업 요약
  - 실제 변경 내용
  - child별 작업 요약 / 반영 내용
  - 테스트 / 검증 결과
  - 남은 리스크 / 제약
  - 후속 개선안 / 다음 추천 작업
  - 참조 artifact / 관련 문서

판정:
- implementation parent final-report 자동 초안 생성은 최신 backend 기준 실전 시나리오에서 정상 동작

## 운영 메모

이번 검증으로 확인된 핵심은 기능 버그보다 **dev backend 재시작 누락 시 오검출이 생길 수 있다**는 점이다.

따라서 backend 변경 후 실전 검증 시에는 반드시:
- 실제 3001 포트 서버가 최신 코드 기준으로 재시작되었는지
- stale go-build cache 실행본이 남아 있지 않은지
를 먼저 확인해야 한다.

## 테스트 task 정리

Playwright smoke 검증에 사용한 테스트 task(`task-363` ~ `task-367`)는 모두 삭제하여 보드를 원상 복구했다.

## 결론

- Phase 1: 실전 검증 통과
- Phase 2: 실전 검증 통과
- 초기 실패처럼 보였던 implementation `final-report` 문제는 코드 결함이 아니라 stale backend 프로세스로 인한 런타임 반영 문제였다.
