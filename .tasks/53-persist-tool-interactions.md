# Task 53: Persist tool interactions in conversation history

Store full tool_use and tool_result messages in the database so the LLM receives complete conversation history — including tool calls, their results, and images — on subsequent turns.

## Context

Currently, when the LLM uses tools during a response, the tool interactions are built in-memory (`internal/api/messages.go:248-295`) but only the final text is persisted (`messages.go:304-316`). On the next user message, `store.ListMessages()` returns plain text messages only, so the LLM loses all context about what tools it called, what data it examined, and any images it saw.

### Key files

**Backend:**
- `internal/store/db.go:65-71` — `messages` table schema: only `role TEXT` and `content TEXT`, no structured content support
- `internal/store/chats.go:14-20` — `Message` struct: `Role`, `Content` (string), no content blocks
- `internal/store/chats.go:106-112` — `CreateMessage`: inserts plain text only
- `internal/store/chats.go:115-134` — `ListMessages`: reads plain text only
- `internal/api/messages.go:97-146` — history reconstruction: converts DB messages to `anthropic.Message` with plain text, no blocks
- `internal/api/messages.go:228-295` — tool loop: appends tool_use/tool_result messages to `req.Messages` in-memory, never persists
- `internal/api/messages.go:304-316` — final persist: stores only `fullText.String()`
- `internal/anthropic/client.go:42-92` — `ContentBlock` with `MarshalJSON`/`UnmarshalJSON` for tool_use, tool_result, image blocks

**What a tool turn actually produces (lost today):**
```
assistant: [{type:"text", text:"Let me check"}, {type:"tool_use", id:"...", name:"read_page", input:{page:3}}]
user:      [{type:"tool_result", tool_use_id:"...", content:"page 3 text..."}]
```

### Relationship to task 51

Task 51 persists user-attached images (region selections). This task persists tool interaction content blocks, which may include images from `snapshot_page` tool results. The image storage approach from task 51 can be reused for tool result images. This task does not depend on task 51 but they share the image persistence concern.

## Checklist

- [x] Migration: add `content_blocks TEXT` column to `messages` table (nullable, stores JSON-serialized content blocks)
- [x] Store: update `Message` struct with optional `ContentBlocks` field
- [x] Store: update `CreateMessage` to persist content_blocks JSON when present
- [x] Store: update `ListMessages` to load content_blocks
- [x] Handler: during tool loop, persist assistant tool_use messages (with any preceding text blocks) to DB
- [x] Handler: during tool loop, persist user tool_result messages to DB
- [x] Handler: when loading history, reconstruct `anthropic.Message` with full content blocks from DB
- [x] Handler: stop storing a duplicate final text-only assistant message when tool iterations occurred (the text is already in the persisted assistant messages)
- [x] Test: store round-trip test for messages with content_blocks
- [x] Test: handler test verifying tool interaction messages survive across turns
- [x] Test: handler test verifying snapshot_page image content persists in tool_result blocks

## Notes

- `ContentBlock.MarshalJSON` already handles serialization of tool_use, tool_result, and image blocks. Add a corresponding `UnmarshalJSON` for deserialization from DB.
- Tool result images (snapshot_page) are base64-encoded PNGs stored inline in content_blocks JSON. This is simple but may bloat the DB. Consider extracting to disk (like task 51) as a follow-up optimization.
- The assistant may emit text before a tool call in the same response. Currently (`messages.go:229-237`) only tool_use blocks are captured — any preceding text block should also be included in the persisted assistant message to match what the API expects.
- Existing conversations in the DB will continue to work — they just won't have content_blocks, so they load as plain text (current behavior).
