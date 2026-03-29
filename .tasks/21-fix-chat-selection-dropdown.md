# Task 21: Fix chat selection dropdown

Consolidate the chat session picker into a single self-contained component. Currently the trigger button and dropdown menu are separate sibling elements in ChatPanel.svelte, with the dropdown rendering far from the button due to positioning issues. Wrap everything (trigger + dropdown + backdrop) in one `position: relative` container so the absolutely-positioned dropdown anchors correctly next to the trigger.

## Context

- **File**: `frontend/src/lib/ChatPanel.svelte`
- Current structure: `<button class="picker-btn">` (with a span label + chevron icon) is inside `.session-picker`, but the `<div class="dropdown">` and `<div class="dropdown-backdrop">` are siblings rendered outside `.session-picker` at the `.chat-panel` level. Since `.chat-panel` doesn't establish the right positioning context, the dropdown ends up at the bottom of the page.
- The dropdown uses `position: absolute; top: 100%` — this would work correctly if its parent were `position: relative` and tightly wrapped the trigger button.
- All dropdown state (`dropdownOpen`, `handleSelect`, `handleBackdropClick`) and per-item delete buttons should remain.

## Checklist

- [x] Extract the session picker (trigger button + dropdown + backdrop) into a single wrapper div with `position: relative`
- [x] Move the `{#if dropdownOpen}` block (backdrop + dropdown div) inside that wrapper so `top: 100%` anchors below the trigger
- [x] Verify dropdown appears directly below the picker button on both desktop and mobile
- [x] Keep existing delete-per-session and new-chat button functionality intact

## Notes

- The fix is purely structural/CSS — move the dropdown markup inside a positioned container, no logic changes needed.
