# Session Chat Inline Message Design

## Summary
- Replace preview-first cards in Sessions > Chat with inline full-text rendering.
- Collapse only messages longer than **12 lines** or **700 characters**.
- Use in-card **더보기 / 접기** instead of the current full-text `View` modal flow.
- Keep attachments and timestamps in the existing card footer.
- Preserve an optional raw payload modal only as a secondary debug affordance.

## Problem
The current UI shows only a 2-line preview and forces users to open a modal for full text. This makes sequential reading of multiple messages slow and interrupts timeline scanning.

## Goals
1. Most messages should be readable without extra clicks.
2. Long messages should remain manageable without exploding card height.
3. The interaction should stay inside the timeline card, not jump into a modal for ordinary reading.

## Proposed UX
- Render message text inline with preserved line breaks.
- If the message exceeds either threshold:
  - show a collapsed in-card version
  - show a `더보기` button
  - allow toggling to `접기`
- Non-long messages render fully with no extra control.
- Keep a small secondary `Raw` action for payload inspection.

## Threshold Rule
A message is collapsible when either condition is true:
- line count > 12
- character count > 700

## Notes
- This is intentionally fixed-rule based, not heuristic.
- Threshold can be tuned later after real usage.
- First iteration prioritizes readability over density.
