# kanban stabilization plan

- Date: 2026-03-24
- Status: Proposed
- Scope: validated kanban orchestration findings를 구현 가능한 phased stabilization plan으로 고정

## 목적

현재 kanban의 핵심 리스크는 단일 버그보다 orchestration hub의 결합도에 있다.

이번 계획은 다음 원칙으로 범위를 고정한다.

1. Phase 1-4는 **behavior-preserving structural stabilization** 만 수행한다.
2. artifact / review policy 의미 변경은 **별도 승인 전까지 미루고**, 구조 분리와 검증 강화만 먼저 한다.
3. frontend/API contract는 의도적으로 바꾸지 않는다.

즉, 이번 문서는 새 정책을 도입하는 문서가 아니라, 이미 검증된 canonical semantics를 깨지 않고 허브 함수를 안전하게 분리하기 위한 실행 순서를 정의한다.

---

## scope lock

### In scope

- `backend/internal/service/kanban/automation_completion.go`
- `backend/internal/service/kanban/planning_execution.go`
- `backend/internal/service/kanban/planning_execution_transition.go`
- `backend/internal/service/kanban/planning_dependency.go`
- `backend/internal/service/kanban/worktree_prepare.go`
- `backend/internal/service/kanban/worktree_base_resolver.go`
- `backend/internal/service/kanban/planning_expansion.go`
- structural phase에서 호출 경계 확인을 위해 읽거나 최소 helper touch가 허용되는 연관 파일
  - `backend/internal/service/kanban/artifact_handoff_apply.go`
  - `backend/internal/service/kanban/artifact_handoff_check.go`
  - `backend/internal/service/kanban/automation_review.go`
  - `backend/internal/service/kanban/automation_execute_git.go`
  - `backend/internal/service/kanban/review_approve.go`
- 필요 시 Phase별 characterization test 추가/정리
  - `backend/internal/service/kanban/artifact_handoff_test.go`
  - `backend/internal/service/kanban/planning_execution_test.go`
  - `backend/internal/service/kanban/worktree_prepare_test.go`
  - `backend/internal/service/kanban/planning_expansion_test.go`
  - `backend/internal/service/kanban/review_actions_test.go`

### Out of scope

- frontend 화면/UX 변경
- REST/SSE/API payload contract 변경
- artifact handoff strictness 재정의
- review approve hard block 정책 변경
- parent/child artifact ownership semantics 변경
- child docs worktree-only 정책의 의미 변경
- cleanup/archive 자동화 추가
- Phase 4 이전 planning stale-retry refactor 단행

---

## 구조 안정화 중 반드시 보존할 canonical semantics

모든 structural phase는 아래 규칙을 깨면 안 된다.

1. finalized artifact owner는 **parent**
2. child output은 **candidate artifact**
3. chaining entry는 **parent only**
4. child docs는 **worktree-only** 정책 유지
5. frontend/API contract는 의도적으로 바꾸지 않음

추가 operational invariant:

- success path에서 기존 event type, status progression, retry gate는 의미를 바꾸지 않는다.
- parent refresh는 현재 count / outcome / stuck summary의 관찰 결과를 유지해야 한다.
- worktree prepare는 parent 공용 준비 모델을 유지해야 한다.
- `completeRunSuccessLocked` 의 review recovery 의미는 유지해야 한다.
  - evidence 부족 시 현재 retry gate / retry count / recovery-triggered restart 의미를 바꾸지 않는다.
- manual approve는 현재 artifact warning을 **soft warning으로만** 다루는 동작을 유지해야 한다.
- parent refresh는 현재 worktree queueing -> review/done transition 연쇄의 관찰 결과를 바꾸지 않는다.
- structural phase에서는 externally visible event type 추가를 기본 금지한다.
  - observability 강화가 필요하면 우선 test assertion / internal helper 경계로 해결한다.

---

## validated findings 요약

1. `completeRunSuccessLocked` 는 success path의 주요 orchestration hub이며, run/task state mutation, artifact sync/mirror, review evidence 판정, handoff validation, parent refresh, activity completion, auto-finish를 한 함수에서 처리한다.
2. `refreshPlanningExecutionForTaskLocked` / `refreshPlanningExecutionLocked` 는 parent refresh hub이며, dependency recomputation, count/outcome/stuck recomputation, event emission, worktree queueing, review transition, auto-done까지 연쇄 수행한다.
3. `worktree_prepare.go` 와 `worktree_base_resolver.go` 는 child execution과 parent doc generation 양쪽에서 재사용되어 blast radius가 크다.
4. `artifact_handoff_*`, `automation_review.go`, `review_approve.go` 는 strict path와 manual approve path의 정책 체감이 완전히 같지 않다.
5. planning expansion stale-retry는 복잡도가 높지만, refactor 기준으로는 아직 부분 검증 상태다. 따라서 먼저 validation phase가 필요하다.

---

## claim matrix

| Claim | Status | Evidence | Planning implication |
| --- | --- | --- | --- |
| run completion success path가 과도하게 큰 허브다 | fact | `backend/internal/service/kanban/automation_completion.go:completeRunSuccessLocked`; `backend/internal/service/kanban/artifact_handoff_test.go:TestCompleteRunSuccessLockedFailsReviewWhenArtifactHandoffMismatchRemains` | Phase 1에서 stage extraction과 characterization test를 먼저 고정 |
| parent refresh가 recompute + transition + side effect를 함께 수행한다 | fact | `backend/internal/service/kanban/planning_execution.go:refreshPlanningExecutionForTaskLocked`, `backend/internal/service/kanban/planning_execution.go:refreshPlanningExecutionLocked`; `backend/internal/service/kanban/planning_execution_test.go:TestRefreshPlanningExecutionLockedMeetingParentAutoDoneWhenChildrenPassed` | Phase 2에서 recompute / transition / side effect 경계를 분리 |
| parent worktree prepare는 child execution과 parent 문서 생성 모두에 영향이 크다 | fact | `backend/internal/service/kanban/worktree_prepare.go:EnsureParentWorktreePrepared`, `backend/internal/service/kanban/worktree_base_resolver.go:resolveTaskWorktreeBaseLocked`; `backend/internal/service/kanban/worktree_prepare_test.go:TestEnsureParentWorktreePrepared_UsesSourceWorktreeLineageWhenArtifactsOnlyExistThere`, `backend/internal/service/kanban/automation_execute_test.go:TestResolveTaskExecutionContext_PreparesParentWorktreeOnDemand` | Phase 3에서 base resolution과 prepare orchestration의 경계를 먼저 고정 |
| artifact handoff, review evidence, manual approve 사이 strictness가 체감상 일관되지 않다 | fact | `backend/internal/service/kanban/artifact_handoff_check.go:evaluateArtifactHandoffLocked`, `backend/internal/service/kanban/automation_review.go:hasReviewEvidenceLocked`, `backend/internal/service/kanban/review_approve.go:validateReviewApproveLocked`; `backend/internal/service/kanban/review_actions_test.go:TestValidateReviewApprove_WarnsWhenFinalDocPathMissing` | 구조 정리는 가능하지만 semantics 정렬은 승인 필요 Phase로 분리 |
| planning expansion stale-retry는 refactor 전에 별도 검증이 필요하다 | partial | `backend/internal/service/kanban/planning_expansion.go:runPlanningExpansion`; existing baseline in `backend/internal/service/kanban/planning_expansion_test.go:TestPlanningExpansionInvalidOutputResetsParentToNotStarted` and `...:TestPlanningExpansionFailsWhenExecutableParentFinalProposalParsingFails` | Phase 4를 validation-first로 두고, 구조 분리는 후속 승인 이후 진행 |
| artifact/review strictness를 완전히 통일할 수 있다 | hypothesis | 현재 근거는 위 strictness 차이 관찰까지이며, 목표 semantics 합의 문서는 아직 없음 | Phase 5에서 approval-needed policy 질문으로만 다룸 |

---

## atomic commit strategy

각 phase는 **작은 테스트 고정 -> 구조 분리** 순서를 기본으로 한다.

### 공통 규칙

1. structural commit과 policy commit을 섞지 않는다.
2. 한 commit은 한 허브만 다룬다.
3. event 이름, status 의미, API contract 변경이 섞이면 phase를 되돌려 쪼갠다.
4. 새 helper 추출은 허용하지만, 의미 변경이 섞이면 안 된다.

### 권장 commit 단위

#### Phase 1
- `test(kanban): pin run completion success-path behavior`
- `refactor(kanban): split completion success stages without policy changes`

#### Phase 2
- `test(kanban): pin planning execution refresh transitions`
- `refactor(kanban): isolate planning execution recompute and side effects`

#### Phase 3
- `test(kanban): pin parent worktree prepare and base resolution`
- `refactor(kanban): isolate worktree prepare orchestration boundaries`

#### Phase 4
- `test(kanban): validate planning expansion stale-retry and relaunch behavior`
- 필요 시 후속 문서 커밋만 추가하고, refactor commit은 승인 전 보류

#### Deferred Phase 5
- policy approval 없이는 코드 커밋 금지

---

## 공통 검증 명령

모든 structural phase는 아래 전체 회귀를 기본으로 수행한다.

```bash
go test ./backend/internal/service/kanban
```

phase별 빠른 검증 예시는 아래를 사용한다.

```bash
go test ./backend/internal/service/kanban -run 'TestCompleteRunSuccessLockedFailsReviewWhenArtifactHandoffMismatchRemains'
go test ./backend/internal/service/kanban -run 'TestCompleteRunFailureLockedExhaustedRetriesMovesChildToReviewWithoutBlocking'
go test ./backend/internal/service/kanban -run 'TestRefreshPlanningExecutionLocked(DirectParentBecomesDone|MeetingParentBecomesReview|MeetingParentAutoDoneWhenChildrenPassed)'
go test ./backend/internal/service/kanban -run 'TestEnsureParentWorktreePrepared_(CreatesMetadataAndBranch|UsesSourceWorktreeLineageWhenArtifactsOnlyExistThere|FallsBackToDevWhenSourceLineageIsUnavailable)'
go test ./backend/internal/service/kanban -run 'TestResolveTaskExecutionContext_(UsesReadyParentWorktree|PreparesParentWorktreeOnDemand|PropagatesPrepareFailure)'
go test ./backend/internal/service/kanban -run 'Test(PlanningExpansionDirectCreatesSingleChild|PlanningExpansionMeetingCreatesMultipleChildrenAndUsesTeamLeaderFallback|PlanningExpansionUsesTaskScopedSessionKey)'
go test ./backend/internal/service/kanban -run 'TestReviewApprove(MovesParentToDone|ClearsReviewDerivedState)'
```

필요 시 artifact/review 회귀를 별도로 묶는다.

```bash
go test ./backend/internal/service/kanban -run 'Test(StartTaskLockedBlocksWhenExpectedArtifactsMissing|EvaluateArtifactHandoffAllowsDescriptorPathWhenSingleArtifactMatchesState|SyncTaskActualArtifactsFromLatestEvidenceLocked)'
```

---

# Phase 1 - run completion success-path structural split

## 목표

`completeRunSuccessLocked` 의 현재 success path를 **관찰 가능한 단계 단위**로 분리한다.

핵심은 아래 흐름을 helper 경계로 나누되, 현재 의미는 그대로 유지하는 것이다.

1. run 종료 기록
2. task review 진입 기본 처리
3. actual artifact sync + markdown mirror
4. review evidence 판정
5. artifact handoff 판정
6. review pass/fail 후속 처리
7. parent refresh + activity completion + auto-finish

## In scope files

- `backend/internal/service/kanban/automation_completion.go`
- 필요 시 호출 경계 정리를 위한 최소 보조 파일
  - `backend/internal/service/kanban/artifact_handoff_apply.go`
  - `backend/internal/service/kanban/artifact_handoff_check.go`
  - `backend/internal/service/kanban/automation_review.go`
- 테스트
  - `backend/internal/service/kanban/artifact_handoff_test.go`

## Non-goals

- review evidence acceptance 기준 변경
- artifact handoff warning 문구 변경
- auto-finish policy 변경
- review failure를 done으로 완화하는 정책 변경

## 구현 포인트

- 함수 내부 mutation과 early return 지점을 named stage로 드러낸다.
- failure/hold reason은 현재 event reason 값을 유지한다.
- `review_failed`, `review_passed`, `run_succeeded` 이벤트 의미는 유지한다.
- helper는 side effect를 숨기지 말고, 이름만으로 단계가 읽히게 만든다.

## 테스트

```bash
go test ./backend/internal/service/kanban -run 'TestCompleteRunSuccessLockedFailsReviewWhenArtifactHandoffMismatchRemains'
go test ./backend/internal/service/kanban -run 'TestCompleteRunFailureLockedExhaustedRetriesMovesChildToReviewWithoutBlocking'
go test ./backend/internal/service/kanban -run 'Test(StartTaskLockedBlocksWhenExpectedArtifactsMissing|SyncTaskActualArtifactsFromLatestEvidenceLocked)'
go test ./backend/internal/service/kanban
```

## 종료 기준

- `completeRunSuccessLocked` 본문이 stage orchestration 수준으로 축소된다.
- success path의 status/event/result는 기존 테스트 기준으로 동일하다.
- artifact sync, evidence, handoff, auto-finish의 원인 추적 지점이 코드상 분리된다.

---

# Phase 2 - planning execution refresh structural split

## 목표

`refreshPlanningExecutionForTaskLocked` / `refreshPlanningExecutionLocked` 를 parent derived-state recompute 허브와 transition/side effect 허브로 분리한다.

핵심은 아래를 나누는 것이다.

1. dependency refresh
2. counts recompute
3. outcome / stuck recompute
4. changed 여부 판정과 event append
5. worktree queueing
6. parent review 전이
7. auto-done

## In scope files

- `backend/internal/service/kanban/planning_execution.go`
- `backend/internal/service/kanban/planning_execution_transition.go`
- `backend/internal/service/kanban/planning_dependency.go`
- 필요 시 summary helper 파일
  - `backend/internal/service/kanban/planning_execution_summary.go`
  - `backend/internal/service/kanban/planning_execution_stuck.go`
- 테스트
  - `backend/internal/service/kanban/planning_execution_test.go`

## Non-goals

- direct/meeting planning product semantics 변경
- auto-review / auto-done 판정 기준 변경
- parent count 필드 이름이나 API 노출 변경

## 구현 포인트

- recompute 결과를 struct로 모으고, mutation 적용은 한 곳에서만 수행한다.
- transition은 recompute 결과를 입력으로 받게 해서, 어디서 parent 상태가 바뀌는지 보이게 한다.
- worktree queueing은 refresh 내부 계산과 분리하되, 호출 타이밍은 유지한다.

## 테스트

```bash
go test ./backend/internal/service/kanban -run 'TestRefreshPlanningExecutionLocked(DirectParentBecomesDone|MeetingParentBecomesReview|MeetingParentAutoDoneWhenChildrenPassed)'
go test ./backend/internal/service/kanban -run 'TestRefreshPlanningExecutionLockedQueuesWorktreePrepareForRunnableChildren'
go test ./backend/internal/service/kanban -run 'Test(AutoReviewGeneratesFinalReportForImplementationParent|AutoReviewDoesNotGenerateFinalReportForDocumentationParent)'
go test ./backend/internal/service/kanban
```

## 종료 기준

- parent derived state 계산과 parent 전이/side effect가 코드상 분리된다.
- count, outcome, stuck summary 회귀가 없다.
- meeting/direct parent의 review/done 전이가 기존 테스트 기준으로 유지된다.

---

# Phase 3 - parent worktree prepare boundary stabilization

## 목표

parent 공용 worktree prepare와 base resolution의 경계를 분리해 blast radius를 낮춘다.

핵심은 다음 두 축을 분리하는 것이다.

1. lineage 기반 base resolution
2. worktree prepare orchestration, state update, async queueing

## In scope files

- `backend/internal/service/kanban/worktree_prepare.go`
- `backend/internal/service/kanban/worktree_base_resolver.go`
- 필요 시 최소 caller touch
  - `backend/internal/service/kanban/planning_execution.go`
  - `backend/internal/service/kanban/automation_execute_git.go`
- 테스트
  - `backend/internal/service/kanban/worktree_prepare_test.go`

## Non-goals

- parent shared worktree 모델 폐기
- child별 독립 worktree 정책 도입
- source lineage 규칙 변경
- docs worktree-only 의미 변경

## 구현 포인트

- base resolution 결과는 pure data 성격으로 다루고, prepare side effect와 분리한다.
- `EnsureParentWorktreePrepared` 는 orchestration entry로 남기되, 실패 원인과 fallback 분기를 helper화한다.
- `resolveParentExecutionWorktree` 와 parent doc generation caller는 기존 contract를 유지한다.

## 테스트

```bash
go test ./backend/internal/service/kanban -run 'TestEnsureParentWorktreePrepared_(CreatesMetadataAndBranch|UsesSourceWorktreeLineageWhenArtifactsOnlyExistThere|FallsBackToDevWhenSourceLineageIsUnavailable)'
go test ./backend/internal/service/kanban -run 'TestResolveTaskExecutionContext_(UsesReadyParentWorktree|PreparesParentWorktreeOnDemand|PropagatesPrepareFailure)'
go test ./backend/internal/service/kanban
```

## 종료 기준

- base resolution reasoning과 prepare 상태 mutation이 코드상 분리된다.
- source worktree lineage 보존, dev fallback, ready metadata 생성이 기존 테스트 기준으로 유지된다.
- child execution / parent doc generation caller의 contract 변화가 없다.

---

# Phase 4 - planning expansion stale-retry validation first

## 목표

`runPlanningExpansion` 의 stale-version retry / unlock-relock / relaunch 흐름은 바로 refactor하지 않고, 먼저 **검증 가능한 모델**로 고정한다.

이 phase의 목적은 refactor가 아니라 다음 질문의 답을 증거로 확보하는 것이다.

1. stale version 감지 시 현재 어떤 relaunch 경로를 타는가
2. 문서 준비 후 재진입에서 어떤 상태가 persisted 되는가
3. retry와 duplicate 느낌의 체감이 실제 중복 실행인지, 정상 relaunch인지

## In scope files

- `backend/internal/service/kanban/planning_expansion.go`
- 필요 시 validation 범위의 보조 파일
  - `backend/internal/service/kanban/planning_expansion_launch.go`
  - `backend/internal/service/kanban/planning_expansion_prepare.go`
  - `backend/internal/service/kanban/intake_orchestration.go`
- 테스트
  - `backend/internal/service/kanban/planning_expansion_test.go`

## Non-goals

- stale-retry loop 단순화 refactor
- planning doc를 제어 흐름에서 제거하는 의미 변경
- planning output normalization policy 변경

## 구현 포인트

- 먼저 characterization test와 traceable helper 경계를 만든다.
- stale retry와 relaunch는 우선 **기존 event / state / error field 범위 안에서** test assertion 가능해야 한다.
- 새 externally visible event type 추가는 이 phase의 기본 수단으로 사용하지 않는다.
- 이 phase 종료 전에는 retry 의미를 바꾸지 않는다.

## 테스트

현재 baseline 검증과 validation-first 진입 게이트는 아래를 사용한다.

```bash
go test ./backend/internal/service/kanban -run 'Test(PlanningExpansionDirectCreatesSingleChild|PlanningExpansionMeetingCreatesMultipleChildrenAndUsesTeamLeaderFallback|PlanningExpansionUsesTaskScopedSessionKey)'
go test ./backend/internal/service/kanban -run 'Test(PlanningExpansionInvalidOutputResetsParentToNotStarted|PlanningExpansionFailsWhenExecutableParentFinalProposalParsingFails|PlanningExpansionMeetingChildrenStayTodoWhenAutoDispatchTicks)'
go test ./backend/internal/service/kanban
```

추가 stale-retry 전용 test는 **이 phase의 첫 커밋에서 먼저 도입**한다. 권장 이름은 아래와 같다.

- `TestPlanningExpansionStaleVersionRelaunchesWithoutDuplicatingChildren`
- `TestPlanningExpansionRetryAfterDocPreparationKeepsSingleExecutionOutcome`

후속 refactor phase는 위 characterization test가 먼저 추가된 뒤에만 진입한다.

## Phase 4 validation result

이번 validation-first 작업으로 아래 현재 동작을 테스트로 고정했다.

- stale check는 `runPlanningExpansion` 의 apply 이전에 **한 번만** 수행된다.
- stale version이 감지되면 `planning_expansion_requeued(reason=task_changed_during_expansion)` 후 즉시 `stale_retry` reason으로 relaunch된다.
- stale retry가 발생해도 active child는 중복 생성되지 않고, 최종 성공 outcome은 1회로 수렴한다.
- `EnsureParentPlanDoc` / `EnsureParentFinalProposalDoc` 단계에서 parent version이 증가해도, 현재 구현에는 **두 번째 stale gate가 없다**.
- stale-retry 관련 event/state evidence는 성공 후 persistence reload 이후에도 유지된다.

추가된 characterization test는 아래 3개다.

- `TestPlanningExpansionStaleVersionRelaunchesWithoutDuplicatingChildren`
- `TestPlanningExpansionRetryAfterDocPreparationKeepsSingleExecutionOutcome`
- `TestPlanningExpansionStaleRetryPersistsEventEvidenceOnReload`

아직 별도 validation이 남아 있는 체크포인트는 하나다.

- `planning_expansion_requeued` 가 persist된 직후, `stale_retry` relaunch 전에 프로세스가 종료되는 경우의 restart semantics

## 종료 기준

- stale-retry와 relaunch가 테스트 또는 명시적 event 단위로 추적 가능하다.
- 중복 실행처럼 보이는 현상에 대해 최소 1개 이상의 재현 가능한 characterization test가 추가된다.
- 구조 분리 필요성은 문서화되지만, semantics 변경 없는 범위만 다음 단계로 넘긴다.

---

# Deferred Phase 5 - policy alignment, approval required

## 목표

artifact handoff, review evidence, manual approve strictness의 정책 정렬 여부를 별도 승인 안건으로 분리한다.

이 phase는 structural stabilization과 섞으면 안 된다.

## Candidate files

- `backend/internal/service/kanban/artifact_handoff_check.go`
- `backend/internal/service/kanban/artifact_handoff_apply.go`
- `backend/internal/service/kanban/artifact_handoff_validation.go`
- `backend/internal/service/kanban/automation_review.go`
- `backend/internal/service/kanban/review_approve.go`

## Deferred questions

1. artifact mismatch는 자동 경로와 수동 approve 경로에서 같은 수준으로 hard block 해야 하는가
2. evidence parsing은 파일 존재/artifact record보다 더 우선하는가, 아니면 보조 신호인가
3. heuristic text evidence 허용 범위를 줄일 것인가
4. manual approve는 warning bypass인지, 별도 signed override인지
5. parent review 진입과 parent done 승인 사이에서 artifact strictness를 어디에 둘 것인가

## 승인 전 비허용 항목

- approve 경로 hard block 기본값 변경
- evidence acceptance 기준 상향/완화
- handoff descriptor path 허용 범위 변경
- artifact mismatch failure reason 재정의

## 종료 기준

- 별도 승인 문서 없이 구현 시작하지 않는다.
- 정책 변경이 필요하면 design/approval 문서를 먼저 만든다.

---

## 권장 실행 순서

1. Phase 1, run completion success-path 분리
2. Phase 2, parent refresh recompute/transition 분리
3. Phase 3, worktree prepare/base resolution 경계 분리
4. Phase 4, planning stale-retry validation 고정
5. 승인 후에만 Deferred Phase 5 진행

이 순서를 권장하는 이유는 다음과 같다.

- Phase 1과 Phase 2가 현재 가장 큰 orchestration hub를 직접 줄인다.
- Phase 3은 blast radius가 넓지만 semantics는 이미 비교적 명확하다.
- Phase 4는 부분 검증 상태라서 refactor보다 validation이 먼저다.
- Phase 5는 구조 문제가 아니라 정책 합의 문제이므로 분리해야 안전하다.

---

## 한 줄 결론

kanban 안정화의 우선순위는 새 정책 도입이 아니라, **현재 의미를 유지한 채 run completion, parent refresh, worktree prepare, planning stale-retry를 단계적으로 분리하고 검증 가능하게 만드는 것**이다.
