# 칸반 시스템 가이드

> `backend/internal/kanban/` 패키지의 전체 구조와 동작을 설명하는 문서.
> 아티팩트 라이프사이클의 상세 규칙은 [kanban-artifact-lifecycle.md](./kanban-artifact-lifecycle.md) 참조.

---

## 1. 패키지 규모

칸반은 이 프로젝트에서 가장 큰 서브시스템이다.

- **파일 수**: ~70개 (테스트 포함)
- **주요 소스**: `store.go` (코어), `automation.go`, `intake_orchestration.go`, `planning_execution.go`, `artifact_handoff.go`, `review_actions.go` 등

---

## 2. 상태 머신

태스크는 6단계 상태를 순차적으로 거친다.

```
intake → planning → todo → in_progress → review → done
```

| 상태 | 의미 | 주요 동작 |
|------|------|-----------|
| `intake` | 원시 입력 | Intake Orchestration (Dominique RPC) |
| `planning` | 계획 수립 | 구조화, 분해, 자식 태스크 생성 |
| `todo` | 실행 대기 | Auto-Dispatch 대기 |
| `in_progress` | 실행 중 | 에이전트가 작업 수행 |
| `review` | 리뷰 대기 | 승인/차단/재오픈 |
| `done` | 완료 | 후속 체이닝 가능 |

### 2.1 상태 전이 규칙

```
intake ──[orchestration]──→ planning
planning ──[expansion]────→ todo (자식 생성 완료)
todo ──[dispatch]─────────→ in_progress
in_progress ──[완료]──────→ review
review ──[approve]────────→ done
review ──[reopen]─────────→ planning 또는 in_progress
review ──[block]──────────→ review (blocked 플래그)
```

- 부모 태스크: 모든 자식 `done` → 자동으로 `review` 전이 (Auto-Finish)
- 자식 태스크: 개별 실행 완료 → `review` 전이

---

## 3. 태스크 분류

태스크는 생성 시 자동으로 분류된다.

### 3.1 TaskKind

| 종류 | 기준 |
|------|------|
| `parent` | `ParentTaskID`가 없는 태스크 |
| `child` | `ParentTaskID`가 있는 태스크 |

### 3.2 TaskCategory

| 카테고리 | 설명 |
|----------|------|
| `implementation` | 코드 구현 |
| `documentation` | 문서 작성 |
| `operations` | 운영/인프라 |
| `unknown` | 분류 불가 |

### 3.3 Priority

`p0` (가장 높음) ~ `p3` (가장 낮음).

---

## 4. 자동화 루프

`StartAutomation(ctx)`이 goroutine으로 실행되며, 주기적으로 태스크 상태를 점검하고 자동 전이를 수행한다.

### 4.1 Intake Orchestration

**파일**: `intake_orchestration.go`

`intake` 상태의 태스크를 Dominique AI에게 RPC로 전달하여 구조화.

```
intake 태스크 → Dominique RPC → 구조화된 Goal, Constraints, DoneCriteria
                              → PlanningMode 결정 (solo/collaborative)
                              → planning 상태로 전이
```

- 타임아웃: `OCL_KANBAN_DOMINIQUE_SUMMON_TIMEOUT` (기본 90초)
- 재시도: `OCL_KANBAN_DOMINIQUE_RETRY_MAX` (기본 1회)
- 실패 시: `OrchestrationStatus = "failed"` + 수동 처리 대기

### 4.2 Planning Execution & Expansion

**파일**: `planning_execution.go`, `planning_expansion.go`

`planning` 상태의 태스크를 분해하여 자식 태스크를 생성.

```
planning 태스크 → Dominique RPC → 계획 문서 생성
                               → 자식 태스크 자동 생성 (dependency 포함)
                               → todo 상태로 전이
```

**Decomposition 상태**:
- `not_started` → `locked` → `splitting` → `completed`
- `manual_review`: 수동 개입 필요 시

### 4.3 Auto-Dispatch

**파일**: `automation_execute.go`

`todo` 상태 태스크를 자동으로 `in_progress`로 전이하여 에이전트에게 실행 위임.

- 라우팅 결정: `routing_decision.go`의 confidence 기반 자동 배정
- 블랙리스트: `routing_config.go`로 특정 에이전트 제외
- 워크트리 준비: `worktree_prepare.go`로 Git 워크트리 생성

### 4.4 Auto-Finish

**파일**: `automation_autofinish.go`

부모 태스크의 모든 자식이 `done`이 되면 자동으로 `review` 전이.

```
자식 A: done ──┐
자식 B: done ──┼──→ 부모: in_progress → review (자동)
자식 C: done ──┘
```

- Artifact Handoff 검증 수행
- 부모 태스크의 `ChildDoneCount` == `ChildCount` 시 트리거

---

## 5. Worktree 관리

태스크 실행 시 Git 워크트리를 생성하여 격리된 환경에서 작업.

### 5.1 Base Resolution

**파일**: `worktree_prepare.go`

워크트리 베이스 브랜치/커밋 결정 우선순위:

1. **Source Worktree**: 소스 태스크의 워크트리가 존재하면 사용
2. **Source Branch**: 소스 태스크의 브랜치 사용
3. **Fallback**: `dev` 브랜치

`WorktreeBaseMode` 필드에 선택 근거를 기록:
- `source_worktree`, `source_branch`, `dev_fallback`

### 5.2 Cleanup Guard

**파일**: `worktree_cleanup_guard.go`

다운스트림 태스크가 참조 중인 워크트리의 삭제를 방지.

- 참조 카운트 기반 (`WorktreeCleanupRefCount`)
- 모든 다운스트림 태스크가 완료되어야 정리 가능
- `WorktreeCleanupState`: `pending`, `deferred`, `cleaned`

---

## 6. 아티팩트 체계

### 6.1 3단계 아티팩트

| 단계 | 필드 | 설명 |
|------|------|------|
| Expected | `ExpectedArtifacts[]` | 계획 시 선언된 예상 산출물 |
| Actual | `ActualArtifacts[]` | 실행 후 실제 생성된 산출물 |
| Finalized | `FinalizedArtifacts[]` | 리뷰 승인 후 확정된 산출물 |

### 6.2 문서 아티팩트

| 문서 | 경로 패턴 | 생성 시점 |
|------|-----------|-----------|
| 계획 문서 | `docs/task-<id>/plan.md` | planning 단계 |
| 최종 제안서 | `docs/task-<id>/final-proposal.md` | documentation 부모 리뷰 승인 |
| 최종 보고서 | `docs/task-<id>/final-report.md` | implementation 부모 리뷰 승인 |

### 6.3 Handoff 검증

**파일**: `artifact_handoff.go`

리뷰 전 아티팩트 완성도를 자동 검증.

`HandoffStatus`:
- `none` → `pending` → `ready` 또는 `warning` 또는 `blocked` 또는 `failed`

`HandoffWarnings[]`: 누락/불일치 아티팩트 목록

### 6.4 체이닝 규칙

**핵심**: 후속 태스크는 **부모의 finalized artifact에서만** 생성 가능.

자식 산출물은 candidate에 불과하며, 부모 리뷰를 거쳐 finalized로 승격되어야 다운스트림에서 사용 가능. 자세한 규칙은 [kanban-artifact-lifecycle.md](./kanban-artifact-lifecycle.md) 참조.

---

## 7. Source Lineage (출처 추적)

**파일**: `source_lineage.go`

후속 태스크 생성 시 원본 태스크의 메타데이터를 저장하여 아티팩트 출처를 추적.

| 필드 | 설명 |
|------|------|
| `SourceTaskID` | 원본 태스크 ID |
| `SourceFinalDocPath` | 원본의 최종 문서 경로 |
| `SourceArtifactPaths[]` | 원본의 모든 아티팩트 경로 |
| `SourceWorktreePath` | 원본의 워크트리 경로 |
| `SourceBranch` | 원본의 Git 브랜치 |
| `SourceBaseCommit` | 원본의 베이스 커밋 |

이 정보를 기반으로 워크트리 베이스를 결정하고, 아티팩트를 올바른 소스에서 참조.

---

## 8. Dominique AI 연동

**파일**: `dominique_summon.go`

Dominique는 오케스트레이션을 담당하는 특수 에이전트. Gateway RPC를 통해 호출.

**소환 정책** (`DominiqueSummonPolicy`):
- `Timeout`: RPC 데드라인 (기본 90초)
- `RetryMax`: 최대 재시도 (기본 1회)
- `SessionKey`: Dominique 세션 키 (기본 `agent:dominique:main`)

**소환 상태** (`DominiqueSummonStatus`):
- `none` → `summoned` → `joined` 또는 `manual_required`

**트리거 사유** (`DominiqueTriggerReasons[]`): 왜 Dominique가 필요한지 기록.

**수동 폴백**: `DominiqueManualFallback = true` 설정 시 사용자가 직접 처리.

---

## 9. 리뷰 액션

**파일**: `review_actions.go`

`review` 상태에서 가능한 액션:

| 액션 | 효과 |
|------|------|
| `approve` | `done` 전이, 아티팩트 finalize |
| `approve-note` | 노트 첨부 후 승인 |
| `reopen-parent` | 부모 재실행 (planning 또는 in_progress로 복귀) |
| `reopen-child` | 특정 자식 재계획/재실행 |
| `block` | 차단 (사유 기록) |

### Block Reasons

**파일**: `block_reasons.go`

| 사유 | 설명 |
|------|------|
| `dependency_waiting` | 의존 태스크 미완료 |
| `artifact_handoff` | 아티팩트 핸드오프 실패 |
| `manual_block` | 사용자 수동 차단 |
| `review_block` | 리뷰 결과 차단 |

---

## 10. Store 영속화

**파일**: `store.go`

### 10.1 인메모리 구조

```go
type Store struct {
    tasks                    map[string]Task
    eventsByTask             map[string][]Event
    commentsByTask           map[string][]Comment
    runsByTask               map[string][]TaskRun
    activitySnapshotByTask   map[string]TaskActivitySnapshot
    activityEventsByTask     map[string][]TaskActivityEvent
    // ...
}
```

### 10.2 파일 영속화

- **경로**: `kanban_store.json` (기본)
- **방식**: 전체 상태를 JSON으로 직렬화 → 파일에 원자적 쓰기
- **Debounce**: `OCL_KANBAN_PERSIST_FLUSH` (기본 0 = 즉시)
- **파일 락**: `OCL_KANBAN_PERSIST_LOCK_TIMEOUT` (기본 2초)
- **동시성 제어**: `Task.Version` 필드로 낙관적 잠금

### 10.3 이벤트 소싱

모든 상태 변경은 `Event` 레코드를 생성하여 감사 추적 가능:

```go
type Event struct {
    ID        string
    TaskID    string
    Type      string   // "status_changed", "comment_added", ...
    Actor     string
    From, To  string   // 상태 변경 시
    Meta      map[string]any
    CreatedAt time.Time
}
```

---

## 11. 라우팅 & 배정

**파일**: `routing_decision.go`, `routing_config.go`

태스크를 적절한 에이전트에게 자동 배정.

| 필드 | 설명 |
|------|------|
| `RoutingTeam` | 대상 팀 |
| `RoutingDecision` | 라우팅 결정 사유 |
| `RoutingConfidence` | 신뢰도 점수 |
| `RoutingSuggestedAssignee` | 추천 에이전트 |
| `RoutingSuggestedPriority` | 추천 우선순위 |

**RoutingConfig**:
- `ConfidenceThreshold`: 자동 배정 최소 신뢰도
- `BlacklistAgentIDs`: 배정 제외 에이전트 목록

---

## 12. Activity 모니터링

**파일**: `activity.go`

실행 중 태스크의 활동을 실시간 추적.

**TaskActivitySnapshot**: 태스크의 현재 상태 스냅샷
- `Status`: `started`, `working`, `testing`, `done`, `failed`, `stale`
- `Stale`: 일정 시간 활동 없으면 stale 표시
- `ElapsedMs`: 경과 시간

**TaskActivityEvent**: 개별 활동 이벤트 기록 (작업 시작, 테스트 실행 등).

---

## 13. Outcome Notification

**파일**: `outcome_notify.go`

태스크 완료/실패 시 알림 전송.

- **Discord**: `OCL_KANBAN_NOTIFY_DISCORD_TARGET` 설정 시 Discord 채널로 전송
- **Session**: `OCL_KANBAN_NOTIFY_SESSION_KEY` 세션으로 메시지 전송 (기본값: `kanban:notify:discord`)
- **자식 배치**: 동일 부모의 자식 완료를 묶어서 일괄 알림

---

## 14. 파일 구조 요약

```
kanban/
├── store.go                          # Task, Store, Event, Comment 정의 + CRUD
├── automation.go                     # 자동화 루프 (StartAutomation, tickAutomation)
├── automation_execute.go             # Auto-Dispatch (todo → in_progress)
├── automation_autofinish.go          # Auto-Finish (자식 완료 → 부모 review)
├── automation_review.go              # Auto-Review 로직
├── intake_orchestration.go           # Intake → Planning (Dominique RPC)
├── planning_execution.go             # Planning → Todo (계획 실행)
├── planning_expansion.go             # 계획 → 자식 태스크 분해
├── planning_dependency.go            # 의존성 상태 관리
├── review_actions.go                 # 리뷰 액션 (approve, block, reopen)
├── routing_decision.go               # 라우팅 결정 로직
├── routing_config.go                 # 라우팅 설정 (DB 기반)
├── artifact_handoff.go               # 아티팩트 핸드오프 검증
├── finalized_artifact.go             # Finalized 아티팩트 관리
├── task_artifact_mirror.go           # 아티팩트 경로 참조 유지
├── source_lineage.go                 # 출처 추적
├── worktree_prepare.go               # Git 워크트리 생성
├── worktree_cleanup_guard.go         # 워크트리 삭제 방지
├── dominique_summon.go               # Dominique AI 소환
├── orchestration_policy.go           # 오케스트레이션 정책 (lock, split)
├── plan_doc.go                       # 계획 문서 생성
├── child_plan_doc.go                 # 자식 계획 문서
├── final_doc.go                      # 최종 문서
├── final_report_doc.go               # 최종 보고서
├── final_proposal_doc.go             # 최종 제안서
├── final_proposal_review_refresh.go  # 제안서 리뷰 갱신
├── final_proposal_work_packages.go   # 작업 패키지
├── plan_preview.go                   # 계획 미리보기
├── activity.go                       # 활동 스냅샷/이벤트
├── block_reasons.go                  # 차단 사유
├── display_assignee.go               # 표시용 담당자
├── executor_rpc.go                   # 실행 RPC 래퍼
├── outcome_notify.go                 # 결과 알림
├── project_context.go                # 프로젝트/워크스페이스 레지스트리
├── task_classification.go            # 태스크 자동 분류
├── task_doc_paths.go                 # 문서 경로 헬퍼
└── *_test.go                         # 각 모듈별 테스트
```
