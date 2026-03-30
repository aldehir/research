# Task 58: Remove chat bubbles from chat UI

Redesign the chat message display to remove bubble styling (backgrounds, rounded corners, max-width constraints) in favor of a flat, full-width layout that maximizes horizontal real estate.

## Context

Current chat UI in `MessageThread.svelte` uses a bubble layout:
- `.message` has `padding: 0.75rem`, `border-radius: var(--radius)`, `max-width: 90%`
- User messages (`.message.user`): purple background (`--color-primary-light`), right-aligned (`align-self: flex-end`)
- Assistant messages (`.message.assistant`): gray background (`--color-bg-tertiary`), left-aligned
- Role labels ("You" / "Assistant") sit above each bubble in small uppercase text
- Thread container uses `display: flex; flex-direction: column; gap: 0.75rem`

Key files:
- `frontend/src/lib/MessageThread.svelte` — message rendering + all bubble CSS
- `frontend/src/lib/ChatPanel.svelte` — parent container
- `frontend/src/lib/theme.css` — `--color-primary-light`, `--color-bg-tertiary` (bubble bg colors)
- `frontend/src/lib/MarkdownRenderer.svelte` — markdown content inside assistant messages

Tool chips (`.tool-chip`) and user attachments (`.user-attachments`) also live inside message bubbles and need to work in the new layout.

## Checklist

- [x] Remove bubble backgrounds, border-radius, max-width, and align-self from `.message`
- [x] Restyle role labels as subtle dividers or inline prefixes to differentiate speakers
- [x] Make messages full-width within the thread container
- [x] Add a light separator (border or spacing) between messages instead of bubble gaps
- [x] Verify tool chips and attachments render correctly in flat layout
- [x] Verify streaming/thinking states still look correct
- [x] Check both light and dark themes
- [x] Check mobile layout

## Notes

- Keep role differentiation clear without bubbles — consider a left-border accent color, bold role name, or subtle background stripe.
- Preserve all existing functionality (markdown, tool chips, attachments, streaming).
- The chat panel is already narrow (360px desktop, 75vw mobile), so every pixel of horizontal space matters.
