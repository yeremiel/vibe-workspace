# Kanban Waterfall Flow Realignment (2026-03-14)

## Goal

Re-align the Kanban parent/child workflow so that planning fully completes before child execution begins, following a waterfall-style assignment model.

## Agreed Flow

1. `INTAKE`
2. `PLANNING`
   - team lead summons planning meeting
   - brainstorming / discussion / scope alignment
   - parent planning artifacts are finalized:
     - `docs/task-{parent}/plan.md`
     - `docs/task-{parent}/final-proposal.md`
3. Branch by task nature:
   - document-only parent: `REVIEW`
   - executable/development parent: child TODO creation + assignment, then parent `IN_PROGRESS`
4. Child execution
   - child reads parent `final-proposal.md`
   - child creates/updates `docs/task-{child}/plan.md`
   - child performs task work
   - child accumulates result / handoff notes in the same `plan.md`
5. Parent `REVIEW`
   - not a document-content review
   - integrated verification / artifact presence / readiness review
6. `DONE`

## Parent Artifact Roles

### `docs/task-{parent}/plan.md`
- planning discussion / meeting notes
- explored options and reasoning
- high-level execution direction
- planning context that explains why the final execution decision was made

### `docs/task-{parent}/final-proposal.md`
- execution-ready planning result
- waterfall assignment source of truth
- should contain fixed execution sections such as:
  - work packages
  - owner / assignee
  - scope
  - deliverables
  - handoff / review criteria

## Child Artifact Role

### `docs/task-{child}/plan.md`
- child-local execution plan
- extracted understanding from parent final proposal
- progress log
- result / handoff notes
- continuity anchor for resume / reassignment / stalled work

## Hard Gates

### Child creation gate
- child TODOs must not be created before parent `final-proposal.md` is ready

### Child execution gate
- child must read parent `final-proposal.md`
- child must create/update `docs/task-{child}/plan.md` before substantial work

### Parent review gate
Parent enters `REVIEW` only when all are satisfied:
- all child tasks are done
- child documents are updated
- required artifacts actually exist
- integrated verification is ready

## Review Meaning

### Parent review
- integrated validation of development/document outputs
- not a late-stage review of planning content

### Child review
- verify the child produced enough documentation, artifacts, and minimum verification evidence
- prefer `REVIEW` over `BLOCKED` when work has moved but polish is incomplete

## Blocked Policy

Use `BLOCKED` only for true execution impossibility, for example:
- missing parent `final-proposal.md`
- assignment scope unclear
- required predecessor artifact missing

Do **not** use `BLOCKED` for minor quality gaps such as:
- document polish 부족
- incomplete explanation
- verification notes that are weak but present

Principle:
- if the task cannot move, block it
- if the task moved, document it and send it to review

## State Mapping

### Parent
- document-only:
  - `INTAKE -> PLANNING -> REVIEW -> DONE`
- development/executable:
  - `INTAKE -> PLANNING -> IN_PROGRESS -> REVIEW -> DONE`

### Child
- `TODO -> IN_PROGRESS -> REVIEW -> DONE`

## Current Implementation Gaps

1. child TODOs are created too early
2. parent `final-proposal.md` is not a hard gate
3. parent planning can finish without the full parent document set
4. parent and child document responsibilities are not cleanly separated
5. parent review semantics are not aligned with integrated verification

## Implementation Priority

### Phase 1
1. require parent `plan.md` + `final-proposal.md` before planning completion
2. restore/add parent final proposal generation in planning
3. block child creation until parent final proposal is ready
4. fix parent state transition after planning

### Phase 2
5. generate child TODOs from parent final proposal work packages
6. require child-local `docs/task-{child}/plan.md`
7. make child prompts use parent final proposal as source of truth

### Phase 3
8. relax blocked rules and use review more aggressively
9. redefine child/parent review gates
10. clean up final document lookup semantics

## Notes

- This document captures the agreed workflow baseline.
- Implementation details may be decided internally as long as they preserve these rules.
- Future final reporting can be layered on top later; it is out of scope for this change.
