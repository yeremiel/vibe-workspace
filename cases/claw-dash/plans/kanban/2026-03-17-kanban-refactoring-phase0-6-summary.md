# Kanban 패키지 리팩토링 Phase 0–6 완료 요약

> 실행일: 2026-03-17
> 대상: `backend/internal/kanban/`
> 커밋 범위: `43532cc..3892cd2` (5 commits)

---

## 배경

`kanban` 패키지가 유기적으로 성장하면서 ~13K LOC(프로덕션), 70개 파일 규모에 도달했다. 설계 문서 없이 기능 추가가 반복되면서 god object, 거대 함수, 코드 중복, 테스트 공백 등 구조적 문제가 축적되었다. 기존 동작을 100% 유지하면서 점진적으로 코드 품질을 개선하는 리팩토링을 Phase 0~6까지 실행했다.

---

## Phase별 실행 요약

### Phase 0: 누락 테스트 추가 (안전망)

기존 코드 변경 없이 테스트만 추가하여 이후 리팩토링의 안전망을 확보했다.

| 신규 테스트 파일 | 커버리지 대상 | 테스트 수 |
|---|---|---|
| `store_core_test.go` | `CreateTask`, `UpdateStatus`, `List`, `Counts` 등 18개 exported 메서드 | 660줄 |
| `automation_core_test.go` | `StartTask`, `RetryTask`, `StopTask` 상태 전이 | 291줄 |
| `routing_config_test.go` | `GetRoutingConfig`, `UpdateRoutingConfig` | 213줄 |
| `planning_dependency_test.go` | 의존성 상태 계산 | 224줄 |

**추가 테스트**: 1,388줄 (+0줄 프로덕션 변경)

### Phase 1: `touchTaskLocked` 헬퍼 도입

54회 반복되던 3줄 패턴(`UpdatedAt` 갱신 + `Version` 증가 + map 저장)을 1개 헬퍼로 통합했다.

```go
func (s *Store) touchTaskLocked(task *Task, now time.Time) {
    task.UpdatedAt = now.Format(time.RFC3339)
    task.Version += 1
    s.tasks[task.ID] = *task
}
```

**영향**: 10개 프로덕션 파일, 순 줄 수 감소

### Phase 2: `automation_execute.go` 파일 분할

1,743줄 단일 파일을 관심사별로 4개 파일로 분리했다. 같은 패키지이므로 함수 시그니처 변경 없음.

| 분리 결과 | 줄 수 | 관심사 |
|---|---|---|
| `automation_execute.go` (잔여) | ~870 | 핵심 실행 흐름 |
| `automation_execute_evidence.go` | 393 | evidence JSON 파싱/검증 |
| `automation_execute_discord.go` | 272 | discord 메시지 파싱 |
| `automation_execute_git.go` | 222 | git 브랜치 resolver |

### Phase 3: `routing_decision.go` 거대 함수 분할 및 파일 분리

`applyRoutingDecisionLocked()` 257줄을 분기별 메서드로 추출하고, 평가/헬퍼 로직을 별도 파일로 분리했다.

| 변경 | Before | After |
|---|---|---|
| `routing_decision.go` | 767줄 | 498줄 |
| `routing_evaluation.go` (신규) | — | 166줄 (평가 로직) |
| `routing_helpers.go` (신규) | — | 149줄 (유틸리티) |

주요 추출:
- `applyRoutingAutoAssignLocked()` — 자동 배정 분기
- `applyRoutingManualQueueLocked()` — 수동 큐 분기
- `applyRoutingBrainstormLocked()` — 브레인스톰 분기
- `routingApplyContext` — 공통 상태 전달 구조체
- `routingEventMeta()` — 이벤트 메타 맵 빌더 (15회 반복 제거)

### Phase 4: 문서 생성 인프라 공통화

8개 doc 파일에서 중복되던 디렉토리 생성 + 파일 쓰기 패턴을 `task_doc_paths.go`에 공통 함수로 추출했다.

| 공통 함수 | 역할 |
|---|---|
| `buildDocWriteTargets()` | 프로젝트 루트 + worktree 기반 쓰기 대상 경로 산출 |
| `writeDocToTargets()` | 복수 대상에 문서 동시 기록 |
| `docMkdirAll` / `docWriteFile` | 테스트 주입 포인트 |

**영향**: `plan_doc.go`, `child_plan_doc.go`, `final_proposal_doc.go`, `final_report_doc.go`, `final_proposal_review_refresh.go`, `task_artifact_mirror.go` — 각 파일의 쓰기 로직이 5줄 이하로 축소

### Phase 5: Task 구조체 서브-구조체 그룹화

122개 평탄 필드를 10개 익명 임베딩 서브-구조체로 논리 그룹화했다. JSON 직렬화 키 불변.

| 서브-구조체 | 필드 수 | 관심사 |
|---|---|---|
| `TaskRoutingState` | 5 | 라우팅 결정/신뢰도 |
| `TaskDominiqueState` | 6 | Dominique 소환 상태 |
| `TaskHandoffState` | 1 | 핸드오프 상태 |
| `TaskPlanDocState` | 3 | 계획 문서 상태 |
| `TaskSourceLineage` | 8 | 소스 태스크 계보 |
| `TaskWorktreeState` | 10 | Worktree 상태 |
| `TaskChildCounts` | 6 | 자식 태스크 카운트 |
| `TaskManualQueueState` | 4 | 수동 큐 상태 |
| `TaskPlanningOutcome` | 1 | 계획 결과 |
| `TaskParentStuck` | 1 | 부모 정체 알림 |

Go 임베딩 특성상 `task.WorktreeStatus` 같은 dot access는 변경 없이 동작하며, struct literal만 임베딩 문법으로 수정했다.

**영향**: 3개 프로덕션 파일 + 10개 테스트 파일

### Phase 6: 오케스트레이션 재시도 패턴 공통화

`intake_orchestration`과 `planning_expansion`에서 중복되던 RPC 호출 + 요약 대기, 알림 전송, 재시도 이벤트 로깅 패턴을 `orchestration_runner.go`로 추출했다.

| 공통 함수 | 역할 |
|---|---|
| `orchestrationRPCCall()` | RPC 호출 + 어시스턴트 요약 수집 |
| `sendOrchestrationNotification()` | 세션 알림 전송 (기존 2개 함수 통합) |
| `logOrchestrationRetry()` | 재시도 이벤트 로깅 |
| `orchestrationIdempotencyKey()` | 멱등성 키 생성 |

**순 감소**: 87줄 (2개 함수의 중복 제거)

---

## 수치 요약

| 지표 | Before | After | 변화 |
|---|---|---|---|
| 프로덕션 LOC | ~13,000 | 13,847 | 구조 개선 (파일 분할로 총량 소폭 증가) |
| 테스트 LOC | ~9,000 | 10,435 | +1,435 (Phase 0 안전망) |
| 프로덕션 파일 수 | 38 | 42 | +4 (분할/추출) |
| 테스트 파일 수 | 34 | 38 | +4 (Phase 0) |
| 통과 테스트 수 | 270+ | 304 | +34 |
| 실패 테스트 수 | 0 | 0 | — |
| 총 diff | — | — | +3,454 / −2,118 줄 |
| 커밋 수 | — | 5 | — |

---

## 신규 파일 목록

| 파일 | Phase | 용도 |
|---|---|---|
| `store_core_test.go` | 0 | Store 기본 메서드 테스트 |
| `automation_core_test.go` | 0 | 자동화 상태 전이 테스트 |
| `routing_config_test.go` | 0 | 라우팅 설정 테스트 |
| `planning_dependency_test.go` | 0 | 의존성 계산 테스트 |
| `automation_execute_evidence.go` | 2 | evidence 파싱 로직 |
| `automation_execute_discord.go` | 2 | discord 메시지 로직 |
| `automation_execute_git.go` | 2 | git 브랜치 resolver |
| `routing_evaluation.go` | 3 | 라우팅 평가 로직 |
| `routing_helpers.go` | 3 | 라우팅 유틸리티 |
| `orchestration_runner.go` | 6 | 오케스트레이션 공통 유틸리티 |

---

## 커밋 이력

```
43532cc refactor(kanban): touchTaskLocked 헬퍼 도입 및 automation_execute 파일 분할
bd92a9c refactor(kanban): routing_decision.go 거대 함수 분할 및 파일 분리
0e06014 refactor(kanban): 문서 생성 인프라 공통화 (Phase 4)
4844204 refactor(kanban): Task 122개 평탄 필드를 10개 embedded sub-struct로 그룹화 (Phase 5)
3892cd2 refactor(kanban): 오케스트레이션 재시도 패턴 공통화 (Phase 6)
```

---

## 미실행 항목

| Phase | 내용 | 사유 |
|---|---|---|
| Phase 7 | Store 접근 계층화 (`getTaskLocked`/`setTaskLocked`) | 25개 파일 169회 치환 대비 실질적 이득이 제한적. Phase 1의 `touchTaskLocked`가 핵심 쓰기 패턴을 이미 통합했으므로, 코스트 대비 효율이 낮아 보류 |

---

## 검증

```bash
cd backend && go build ./internal/kanban/...    # 빌드 통과
cd backend && go vet ./internal/kanban/...      # 정적 분석 통과
cd backend && go test ./internal/kanban/... -count=1  # 304 PASS, 0 FAIL
```
