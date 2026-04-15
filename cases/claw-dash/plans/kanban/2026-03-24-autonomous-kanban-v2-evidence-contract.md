# Autonomous Kanban v2 - Executor Evidence Contract

- Date: 2026-03-24
- Status: Draft
- Scope: self-closing kanban을 위해 executor 결과를 machine-verifiable contract와 canonical evidence record로 고정

## 배경

현재 kanban의 child execution은 자동화되어 있지만, completion은 아직 완전히 self-closing 하지 않다.

주된 이유는 executor 결과가 단일 machine contract가 아니라 다음이 섞인 형태이기 때문이다.

- run outcome envelope
- legacy execution evidence JSON
- adapter result text
- legacy template text
- review recovery restart
- auto-finish용 문자열 heuristic

현재 구현에서 evidence의 실제 진실원은 별도 machine record가 아니라 **executor comment 로그**다.

1. executor가 assistant history에 결과를 남긴다.
2. runtime이 token-matched JSON을 추출한다.
3. 그 JSON 문자열이 executor comment로 저장된다.
4. review / artifact handoff / auto-finish가 다시 comment를 스캔한다.

이 구조에서는 “실행 성공”과 “완료 가능한 성공”이 같은 뜻이 아니다. 따라서 Autonomous Kanban v2에서는 executor 결과를 **기계가 직접 review / handoff / auto-finish / parent progression을 판정할 수 있는 단일 contract**로 재정의한다.

## 목표

이 문서의 목표는 다음 하나다.

> executor의 결과를 사람이 읽는 로그가 아니라, kanban automation이 직접 신뢰할 수 있는 structured evidence contract로 고정한다.

이 contract가 만족되면:

1. child review pass/fail
2. artifact handoff 판정
3. auto-finish gate
4. parent progression
5. parent auto-close

를 모두 같은 evidence source로 판정할 수 있어야 한다.

## Non-goals

- planner contract를 여기서 정의하지 않는다.
- artifact ownership semantics를 바꾸지 않는다.
- parent review/manual approve 정책 자체를 여기서 바로 제거하지 않는다.
- worktree/visibility contract 전체를 여기서 확정하지 않는다.
- executor transport 자체를 여기서 확정하지 않는다.

## 유지할 current canonical semantics

이 contract는 아래 규칙을 깨면 안 된다.

1. child output은 candidate artifact다.
2. finalized artifact owner는 parent다.
3. downstream chaining은 parent-finalized artifact에서만 시작한다.
4. worktree isolation은 유지한다.
5. lineage-aware source selection은 유지한다.

즉, executor evidence contract는 **자동 종료를 위한 기반 계약**이지, artifact lifecycle 규칙을 대체하는 문서가 아니다.

## 현재 v1 evidence 흐름

현재 evidence 흐름은 대략 아래와 같다.

1. `automation_execution_prompt.go`가 executor에게 `execToken/executed/evidence/risks` 중심 JSON을 요구한다.
2. `automation_execute.go`가 assistant history에서 latest token-matched summary를 찾는다.
3. `automation_execute_evidence.go`가 JSON을 validate하고, 실패하면 self-repair를 한 번 시도한다.
4. 최종 evidence JSON 문자열은 executor comment로 저장된다.
5. `automation_review.go`는 latest `execution started:` 이후의 executor comment를 다시 스캔해서 review evidence를 판정한다.
6. `artifact_handoff_apply.go`는 같은 evidence에서 `ActualArtifacts`를 추출한다.
7. `automation_completion.go`는 evidence acceptance → artifact handoff → review pass/fail 순서로 child 상태를 닫는다.
8. `automation_autofinish.go`는 evidence를 다시 읽어 test/diff/risk gate를 적용한다.

즉 v1은 다음 구조다.

> chat/history → parse → executor comment 저장 → review/handoff/autofinish가 comment를 재해석

## 현재 구조의 핵심 문제

### 1. success 의미가 둘로 갈라져 있다

- `executeRunWithRPC`는 valid JSON evidence가 있으면 success로 종료할 수 있다.
- 하지만 auto-finish는 다시 `tests passed`, `diff clean`, `no blockers` 같은 문자열 heuristic을 찾는다.

즉 success 판단과 closure 판단이 같은 contract를 쓰지 않는다.

### 2. evidence schema가 너무 약하다

현재 필수는 사실상 다음뿐이다.

- `execToken`
- `executed`
- `evidence`
- `risks`

이 정도로는 acceptance/test/scope/artifact를 기계적으로 닫기 어렵다.

### 3. 진실원이 chat/comment scraping에 의존한다

- JSON object extraction은 generic assistant text에서 candidate를 뽑는다.
- latest parseable message를 선택한다.
- review는 dedicated record가 아니라 comment scan에 의존한다.

즉 evidence source가 deterministic machine channel이 아니다.

### 4. 부족한 evidence를 recovery/retry로 메운다

- invalid evidence JSON은 self-repair prompt를 한 번 더 보낸다.
- insufficient evidence는 review recovery rerun을 한 번 허용한다.

이건 곧 시스템이 primary evidence contract를 아직 충분히 신뢰하지 못한다는 뜻이다.

### 5. legacy / heuristic 성공 경로가 아직 남아 있다

현재 review는 다음도 성공 증적으로 받아들인다.

- legacy execution evidence JSON
- adapter result text (`EXEC_TOKEN=...`, `Result: sent|reacted`)
- 일부 legacy template text

이 구조는 migration에는 유용하지만 self-closing의 최종 상태로는 부적합하다.

## v2 설계 원칙

1. evidence는 정확히 **하나의 JSON object**여야 한다.
2. evidence는 현재 run과 `execToken` / `runId`로 1:1 매칭돼야 한다.
3. review / handoff / auto-finish / parent progression은 모두 같은 canonical evidence record만 읽어야 한다.
4. free-form summary는 사람이 읽는 보조 로그일 뿐, 판정 근거가 아니다.
5. self-repair / review-recovery는 정상 경로가 아니라 명시적 failure handling이어야 한다.

## v2 contract는 두 층이다

Autonomous Kanban v2에서 evidence contract는 schema 하나로 끝나지 않는다.

### 1. Payload Contract

무슨 정보를 반드시 담아야 하는가.

### 2. Channel Contract

그 정보를 어디에 어떻게 저장하고 읽을 것인가.

현재 문제의 절반은 필드 부족이고, 나머지 절반은 **추출 방식**이다. 따라서 두 층을 함께 고정해야 한다.

## v2 Payload Contract

### 목표 상태

executor 결과는 사람이 읽는 prose가 아니라 아래 정보를 가진 구조화 payload여야 한다.

- 실행 자체가 성공했는가
- evidence가 충분한가
- acceptance를 통과했는가
- 어떤 verification을 돌렸는가
- 어떤 artifacts가 생겼는가
- 어떤 risks가 남았는가
- 실패했다면 어떤 failure taxonomy에 속하는가

### Draft schema

```json
{
  "execToken": "tk-123",
  "runId": "run-123",
  "status": "success",

  "executionSucceeded": true,
  "evidenceSufficient": true,
  "acceptancePassed": true,

  "executed": [
    "updated backend/internal/service/kanban/automation_review.go"
  ],

  "verification": {
    "tests": [
      {
        "name": "go test ./backend/internal/service/kanban",
        "status": "passed"
      }
    ],
    "build": [
      {
        "name": "go test compile",
        "status": "passed"
      }
    ],
    "scopeCheck": {
      "status": "passed",
      "notes": [
        "no unrelated file changes"
      ]
    }
  },

  "artifacts": [
    {
      "path": "backend/internal/service/kanban/automation_review.go",
      "state": "modified"
    }
  ],

  "risks": [
    {
      "level": "none",
      "message": "none"
    }
  ],

  "failure": null
}
```

### Required fields

다음 필드는 필수다.

- `execToken`
- `runId`
- `status`
- `executionSucceeded`
- `evidenceSufficient`
- `acceptancePassed`
- `executed`
- `verification`
- `artifacts`
- `risks`

### Field semantics

#### `execToken`
- 현재 run을 식별하는 고유 토큰
- 정확히 일치해야 함
- mismatch는 즉시 invalid contract

#### `runId`
- kanban 내부 run record와 직접 연결되는 식별자
- event / retry / audit correlation에 사용

#### `status`
허용값 예시:
- `success`
- `failed`
- `blocked`
- `aborted`

#### `executionSucceeded`
- executor가 작업 수행 자체를 마쳤는지
- false면 이후 필드가 있어도 closure 불가

#### `evidenceSufficient`
- 결과를 판정할 만큼 충분한 근거가 있는지
- false면 recovery가 아니라 contract failure로 처리하는 방향을 목표로 한다

#### `acceptancePassed`
- done criteria / spec 기준을 충족했는지
- execution success와 별개다
- 작업은 했지만 acceptance 미통과일 수 있다

#### `executed`
- 실제 수행 내용을 요약한 structured list
- free-form prose가 아니라 action summary여야 한다

#### `verification`
auto-finish가 문자열 검색을 하지 않도록 구조화한다.

##### `verification.tests`
- 어떤 테스트/검증을 돌렸는지
- 각각 `passed|failed|skipped`

##### `verification.build`
- 빌드/타입체크/린트 같은 검증
- 각각 `passed|failed|skipped`

##### `verification.scopeCheck`
- 작업 범위 오염 여부
- `passed|failed`
- note 포함 가능

#### `artifacts`
- 실제 변경/생성/삭제/존재 확인된 artifact
- artifact handoff와 actual artifact sync의 단일 입력으로 사용
- current `TaskArtifact`와 가능한 한 정렬한다

#### `risks`
- free-form `none` 한 줄 대신 structured list
- 최소한 `level`과 `message`가 필요하다
- `none`, `low`, `medium`, `high` 수준을 기본으로 본다

#### `failure`
실패 시 원인을 구조적으로 담는다.

예:

```json
{
  "code": "acceptance_failed",
  "message": "integration test did not pass"
}
```

## v2 Channel Contract

### 목표 상태

evidence는 더 이상 generic chat history나 executor comment scan에서 재구성하지 않는다.

원하는 파이프라인은 다음이다.

> executor result → validate → canonical evidence record 저장 → review/handoff/autofinish가 record만 읽음

즉 v2는:

- **v1**: chat/comment가 진실원
- **v2**: evidence record가 진실원, chat/comment는 증빙 UI

### Evidence Record requirements

run 종료 시점에 정확히 하나의 canonical evidence record를 저장한다.

예시 schema:

```json
{
  "taskId": "task-41",
  "runId": "run-41",
  "execToken": "tk-41",
  "source": "agent_chat",
  "collectedAt": "2026-03-24T12:00:00Z",
  "validationStatus": "valid",
  "validationErrors": [],
  "payload": {
    "execToken": "tk-41",
    "runId": "run-41",
    "status": "success",
    "executionSucceeded": true,
    "evidenceSufficient": true,
    "acceptancePassed": true,
    "executed": ["updated backend/service.go"],
    "verification": {
      "tests": [{"name": "go test ./...", "status": "passed"}],
      "build": [],
      "scopeCheck": {"status": "passed", "notes": []}
    },
    "artifacts": [{"path": "backend/service.go", "state": "modified"}],
    "risks": [{"level": "none", "message": "none"}],
    "failure": null
  },
  "rawAssistantSummary": "{...}",
  "recordVersion": 1
}
```

### Read rules

#### Review
- latest valid evidence record만 읽는다.
- comment/body parsing은 판정 입력으로 사용하지 않는다.

#### Artifact handoff
- latest valid evidence record의 `payload.artifacts[]`만 읽는다.
- text fallback extraction은 migration 이후 제거 대상이다.

#### Auto-finish
- evidence record의 `verification`, `risks`, `artifacts`만 읽는다.
- `tests passed`, `diff clean`, `no blockers` 같은 문자열 heuristic은 사용하지 않는다.

#### Parent progression
- raw executor output이 아니라, child의 evidence-record-derived state만 본다.

### Write rules

1. run 종료 시 runtime이 raw assistant summary를 받는다.
2. payload parser가 strict contract 검증을 수행한다.
3. 검증이 통과하면 canonical evidence record를 저장한다.
4. 검증이 실패하면 invalid evidence record 또는 failure metadata를 저장한다.
5. comment/log는 사람용 trace로 남기되, machine truth는 evidence record 한 곳으로 제한한다.

## Failure taxonomy

v2에서 최소한 아래 코드는 구분되어야 한다.

- `execution_failed`
- `invalid_evidence_contract`
- `insufficient_evidence`
- `acceptance_failed`
- `artifact_contract_failed`
- `visibility_contract_failed`
- `blocked_by_policy`

## Review / auto-finish rules

### Child review

child review pass 후보는 다음을 만족해야 한다.

- `executionSucceeded=true`
- `evidenceSufficient=true`
- `acceptancePassed=true`

### Auto-finish

auto-finish는 더 이상 문자열 heuristic을 찾지 않는다. 대신 다음만 본다.

- `verification.tests`
- `verification.build`
- `verification.scopeCheck`
- `risks`
- `artifacts`

### Artifact handoff

artifact handoff는 evidence payload의 `artifacts[]`를 우선 진실원으로 사용한다.

## Migration phases

### Phase A
- run outcome envelope 우선 유지
- legacy JSON / adapter text는 하위 호환 허용
- 단, 새 executor는 v2 payload 출력 강제
- runtime은 evidence record 저장을 먼저 도입

### Phase B
- review path에서 legacy 성공 인정 제거
- self-repair 의존 축소
- comment scan 대신 evidence record 기반 read 전환

### Phase C
- review / handoff / auto-finish 모두 v2 contract only
- legacy parser와 keyword gate 제거

## 금지할 것

Autonomous Kanban v2의 목표 상태에서는 아래를 기본 성공 근거로 인정하지 않는다.

- adapter result text만으로 success 판정
- `tests passed` 같은 plain text만으로 auto-finish 통과
- legacy template prose를 success evidence로 간주
- review recovery restart를 정상 경로로 사용
- generic chat history scraping을 최종 진실원으로 사용

## 예상 효과

이 contract가 도입되면:

1. execution success와 review success의 기준이 하나로 정리된다.
2. auto-finish가 문자열 heuristic에서 벗어난다.
3. manual approve는 기본 경로가 아니라 예외 경로가 된다.
4. self-closing automation에 필요한 첫 번째 machine contract가 생긴다.
5. evidence provenance가 chat/comment가 아니라 runtime-owned record로 이동한다.

## 구현 우선순위

1. `automation_execute_evidence.go`에 v2 payload schema와 strict parser 추가
2. run 종료 시 canonical evidence record 저장 경로 설계/추가
3. `automation_review.go`를 evidence record 중심으로 재구성
4. `artifact_handoff_apply.go`를 `artifacts[]` 중심으로 재정렬
5. `automation_autofinish.go`의 keyword heuristic 제거
6. 그 다음 parent auto-close 조건을 다시 설계

## 결정 문장

Autonomous Kanban v2는 executor evidence를 다음처럼 취급한다.

> executor evidence는 사람이 읽는 참고 로그가 아니라,
> kanban runtime이 validate하여 canonical evidence record로 고정하고,
> review / handoff / auto-finish / closure를 판정하는 단일 machine contract다.
