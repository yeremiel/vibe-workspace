# Autonomous Kanban v2 - Executor Evidence Gap Analysis

- Date: 2026-03-24
- Status: Draft
- Scope: current kanban implementation과 `Autonomous Kanban v2 - Executor Evidence Contract` 사이의 구체적 파일 단위 갭 정리

## Purpose

이 문서는 `docs/plans/2026-03-24-autonomous-kanban-v2-evidence-contract.md` 에서 정의한 목표 상태와 현재 구현 사이의 차이를 **구현 가능한 단위로 분해**하기 위한 문서다.

핵심 질문은 두 가지다.

1. 현재 evidence v1은 어디서 v2와 어긋나는가?
2. 어떤 순서로 바꿔야 self-closing automation에 필요한 evidence contract를 실제 코드에 심을 수 있는가?

## 기준 문서

- Target: `docs/plans/2026-03-24-autonomous-kanban-v2-evidence-contract.md`
- Canonical lifecycle baseline: `docs/architecture/kanban-artifact-lifecycle.md`

## 현재 v1 사실 요약

현재 evidence의 실제 진실원은 canonical evidence record가 아니라 **executor comment 로그**다.

현재 파이프라인:

1. executor가 assistant history에 결과를 남긴다.
2. `automation_execute.go`가 latest token-matched summary를 찾는다.
3. `automation_execute_evidence.go`가 v1 JSON을 validate하고 필요 시 self-repair를 시도한다.
4. `automation_execute_apply.go`가 최종 JSON 문자열을 executor comment로 저장한다.
5. `automation_review.go`가 comment scan으로 review evidence를 판정한다.
6. `artifact_handoff_apply.go`가 comment에서 artifacts를 다시 뽑는다.
7. `automation_autofinish.go`가 comment-derived evidence를 다시 읽어 heuristic gate를 적용한다.

즉, 현재 구조는 다음과 같다.

> chat/history → parse → comment 저장 → review/handoff/autofinish가 comment를 재해석

v2는 이 구조를 다음으로 바꾸고자 한다.

> executor result → validate → canonical evidence record 저장 → review/handoff/autofinish가 record만 읽음

## Gap 분류

이 문서는 갭을 네 묶음으로 나눈다.

1. Payload Contract gaps
2. Channel Contract gaps
3. Read Rule gaps
4. Migration blocker gaps

우선순위는 `P0`(반드시 먼저), `P1`(1차 전환 직후), `P2`(후속 정리)로 표기한다.

---

## 1. Payload Contract gaps

### Gap 1 — parser가 아직 v1 schema만 인정한다

- Priority: **P0**
- 현재 파일:
  - `backend/internal/service/kanban/automation_execute_evidence.go`
- 현재 상태:
  - 필수 필드는 사실상 `execToken`, `executed`, `evidence`, `risks`
  - `artifacts`는 optional
- v2 목표:
  - `runId`
  - `status`
  - `executionSucceeded`
  - `evidenceSufficient`
  - `acceptancePassed`
  - `verification`
  - structured `risks`
  - structured `failure`
  를 포함한 payload를 strict parse/validate 해야 한다.
- 의미:
  - 현재 parser는 “값이 채워져 있다”는 사실은 확인하지만, “자동 종료 가능한 근거가 충분한가”는 표현하지 못한다.

### Gap 2 — executor prompt가 아직 v1 payload를 요구한다

- Priority: **P0**
- 현재 파일:
  - `backend/internal/service/kanban/automation_execution_prompt.go`
- 현재 상태:
  - executor에게 `execToken/executed/evidence/risks` 중심 JSON을 요구
  - success-gate guidance도 prose 기반
- v2 목표:
  - strict payload schema를 prompt 레벨에서도 요구
  - verification/tests/build/scopeCheck 구조를 명시
- 의미:
  - parser를 바꿔도 prompt가 v1을 계속 요구하면 contract migration이 일어나지 않는다.

### Gap 3 — risks / verification semantics가 아직 natural-language 의존이다

- Priority: **P0**
- 현재 파일:
  - `backend/internal/service/kanban/automation_autofinish_gate.go`
- 현재 상태:
  - `tests passed`, `diff clean`, `no blockers` 같은 문자열 검색으로 판정
- v2 목표:
  - `verification.tests`
  - `verification.build`
  - `verification.scopeCheck`
  - structured `risks`
  만 읽는다.
- 의미:
  - payload schema를 확장해도 downstream gate가 문자열 heuristic이면 v2 계약은 사실상 무력하다.

---

## 2. Channel Contract gaps

### Gap 4 — canonical evidence record entity가 존재하지 않는다

- Priority: **P0**
- 현재 파일:
  - `backend/internal/model/kanban/entities.go`
- 현재 상태:
  - comment / run / task entity는 있지만 evidence-record entity가 없다.
- v2 목표:
  - taskId / runId / execToken / source / collectedAt / validationStatus / validationErrors / payload / rawAssistantSummary / recordVersion 등을 가진 evidence record가 필요하다.
- 의미:
  - payload contract를 정의해도 저장할 canonical container가 없으면 truth source 전환이 불가능하다.

### Gap 5 — store에 evidence record index/map이 없다

- Priority: **P0**
- 현재 파일:
  - `backend/internal/service/kanban/store.go`
- 현재 상태:
  - `commentsByTask`, `runtimeByRunID`, `runsByTask`는 있지만 evidence record 저장소가 없다.
- v2 목표:
  - run 기준 / task 기준으로 latest valid evidence record를 읽을 수 있는 store structure 필요
- 의미:
  - review / handoff / auto-finish reader migration의 선행 조건이다.

### Gap 6 — persistence layer가 evidence record를 저장하지 않는다

- Priority: **P0**
- 현재 파일:
  - `backend/internal/service/kanban/store_persistence.go`
- 현재 상태:
  - comments와 runs는 snapshot에 들어가지만 evidence record는 없다.
- v2 목표:
  - evidence record persist / restore 지원
- 의미:
  - 메모리에서만 존재하면 restart 이후 self-closing semantics가 깨진다.

### Gap 7 — run completion이 canonical record 대신 comment를 쓴다

- Priority: **P0**
- 현재 파일:
  - `backend/internal/service/kanban/automation_execute.go`
  - `backend/internal/service/kanban/automation_execute_apply.go`
- 현재 상태:
  - parse 성공 후 `result.Summary`에 JSON 문자열을 담고, success 시 executor comment로 append
- v2 목표:
  - run completion에서 validated payload를 canonical evidence record로 저장하고, comment는 보조 trace로만 남긴다.
- 의미:
  - 지금은 write path부터 comment-first다.

---

## 3. Read Rule gaps

### Gap 8 — review가 evidence record가 아니라 comment scan을 읽는다

- Priority: **P0**
- 현재 파일:
  - `backend/internal/service/kanban/automation_review.go`
- 현재 상태:
  - latest `execution started:` 이후 executor comments를 스캔
  - run outcome envelope, legacy JSON, adapter result text, legacy template를 success 후보로 인정
- v2 목표:
  - latest valid evidence record only
  - legacy success path 제거
- 의미:
  - 현재 review truth가 가장 직접적으로 v2와 충돌하는 부분이다.

### Gap 9 — artifact handoff가 evidence record가 아니라 comment parse/fallback text를 읽는다

- Priority: **P0**
- 현재 파일:
  - `backend/internal/service/kanban/artifact_handoff_apply.go`
- 현재 상태:
  - latest executor comment parse
  - token이 없으면 text fallback 허용
- v2 목표:
  - canonical evidence record의 `payload.artifacts[]`만 읽는다.
- 의미:
  - artifact contract를 strict하게 만들려면 text fallback을 걷어내야 한다.

### Gap 10 — auto-finish가 comment-derived evidence를 다시 읽는다

- Priority: **P0**
- 현재 파일:
  - `backend/internal/service/kanban/automation_autofinish.go`
  - `backend/internal/service/kanban/automation_autofinish_gate.go`
- 현재 상태:
  - latest token-matched evidence JSON을 comments에서 다시 찾는다.
  - 그 후 keyword heuristics로 tests/diff/risks를 판정한다.
- v2 목표:
  - canonical evidence record에서 `verification`, `risks`, `artifacts`만 읽는다.
- 의미:
  - auto-finish는 self-closing의 마지막 게이트라서 reader migration 우선순위가 높다.

### Gap 11 — dependency context도 comments에 기대고 있다

- Priority: **P1**
- 현재 파일:
  - `backend/internal/service/kanban/automation_execution_dependency.go`
- 현재 상태:
  - downstream dependency context가 executor comments 요약에 기대고 있음
- v2 목표:
  - evidence-record-derived summary 또는 explicit handoff summary 사용
- 의미:
  - child → child / child → parent 맥락 전달도 결국 comment truth에서 벗어나야 한다.

---

## 4. Migration blocker gaps

### Gap 12 — self-repair가 정상 evidence flow에 포함돼 있다

- Priority: **P1**
- 현재 파일:
  - `backend/internal/service/kanban/automation_execute_evidence.go`
  - `backend/internal/service/kanban/automation_execute.go`
- 현재 상태:
  - invalid evidence JSON이면 self-repair prompt를 보내 다시 evidence를 받는다.
- v2 목표:
  - self-repair는 명시적 failure handling 또는 운영 보조 도구로만 남기고, 정상 success path에서는 제거
- 의미:
  - primary contract가 불안정하다는 신호를 정상 흐름이 흡수하고 있다.

### Gap 13 — review_recovering rerun이 evidence 부족의 기본 보정 경로다

- Priority: **P1**
- 현재 파일:
  - `backend/internal/service/kanban/automation_completion.go`
- 현재 상태:
  - `insufficient_execution_evidence`이면 1회 `review_recovering` rerun
- v2 목표:
  - evidence 부족은 contract failure로 먼저 드러나야 하며, rerun은 별도 정책이어야 한다.
- 의미:
  - self-closing system에서 증거 부족을 재실행으로 기본 보정하면 contract 신뢰가 계속 약해진다.

### Gap 14 — parent progression이 evidence-derived state를 간접적으로만 본다

- Priority: **P1**
- 현재 파일:
  - `backend/internal/service/kanban/automation_completion.go`
  - `backend/internal/service/kanban/planning_execution_summary.go`
- 현재 상태:
  - parent auto-done은 child `StatusDone` + `AutoReviewPassed`를 보며, evidence record 자체를 직접 참조하지 않는다.
- v2 목표:
  - child closure state가 evidence record 기반으로 산출되도록 먼저 정렬
- 의미:
  - evidence record가 없으면 parent auto-close를 evidence-driven하게 재설계하기 어렵다.

---

## 우선순위 정리

### First wave (P0)

1. v2 payload parser 도입
2. executor prompt를 v2 payload 기준으로 변경
3. evidence record entity/model/store/persistence 도입
4. run completion write path를 evidence-record-first로 전환
5. review reader를 evidence record only로 전환
6. artifact handoff reader를 evidence record only로 전환
7. auto-finish reader를 evidence record + structured verification 기준으로 전환

### Second wave (P1)

1. self-repair를 정상 경로에서 제거
2. review_recovering rerun을 기본 evidence 보정 경로에서 제거
3. dependency summary / parent progression을 evidence-record-derived state 중심으로 전환

### Later cleanup (P2)

1. legacy text success path 제거
2. comment-based evidence tests를 record-based tests로 재작성
3. UI에서 comment/log와 canonical evidence record를 구분해 노출

## 권장 구현 순서

실제 구현 순서는 아래가 가장 안전하다.

1. `automation_execute_evidence.go`
   - v2 payload parser / validator 추가
2. `entities.go` + `store.go` + `store_persistence.go`
   - evidence record model / store / persistence 추가
3. `automation_execute.go` + `automation_execute_apply.go`
   - validated payload를 canonical evidence record로 저장
4. `automation_review.go`
   - review를 comment scan에서 evidence-record-only로 전환
5. `artifact_handoff_apply.go`
   - `payload.artifacts[]`만 사용하도록 전환
6. `automation_autofinish.go` + `automation_autofinish_gate.go`
   - structured verification 기반으로 전환
7. `automation_completion.go`
   - self-repair / review_recovering normal path 제거 또는 축소

## 핵심 결론

현재 evidence v1과 v2 사이의 가장 큰 차이는 parser 수준이 아니다.

> **가장 큰 차이는 truth source가 여전히 executor comments라는 점이다.**

따라서 Autonomous Kanban v2로 가기 위한 첫 번째 실질 작업은 “더 좋은 JSON”만 만드는 것이 아니라,

> **runtime-owned canonical evidence record를 model/store/persistence에 도입하고, 모든 reader를 그쪽으로 옮기는 것**

이다.
