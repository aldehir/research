# Task 60: Full-screen mobile chat and vertical-optimized LLM formatting

Make the chat panel consume the entire screen on mobile (instead of 75vw slide-over) with a close button, and update the system prompt to instruct the LLM to avoid wide horizontal tables and optimize formatting for narrow/vertical screens.

## Context

**Mobile chat panel** — currently a 75vw slide-over overlay in `frontend/src/routes/+layout.svelte`:
- `.chat-overlay-wrapper` has `width: 75vw; min-width: 280px`
- Slides in from right with `transform: translateX(100%)`
- Backdrop overlay behind it closes panel on click
- Mobile state managed by `frontend/src/lib/mobile-layout.svelte.ts` (`toggleChat()`, `closePanel()`)
- `ChatPanel.svelte` hides the collapse button on mobile already (`{#if !getIsMobile()}`)

**System prompt** — defined in `internal/chat/prompt.go`:
- `basePrompt` constant contains all LLM guidelines
- `BuildSystemPrompt(ctx PromptContext)` appends document metadata
- Tests in `internal/chat/prompt_test.go` verify prompt construction
- Used in `internal/api/messages.go` when building chat requests

## Checklist

- [x] Change mobile chat overlay from 75vw to 100vw full-screen
- [x] Add a close/back button to the chat header on mobile
- [x] Remove backdrop overlay since chat is full-screen (or keep as no-op)
- [x] Update system prompt to instruct LLM to use vertical-friendly formatting (avoid wide tables, prefer lists/prose)
- [x] Add test for new prompt content
- [x] Verify mobile chat open/close flow works

## Notes

- The close button should be visible and easy to tap (44px touch target minimum).
- Consider using an arrow-left or X icon for the close action.
- System prompt addition should be concise — one or two sentences about preferring vertical/narrow formatting.
