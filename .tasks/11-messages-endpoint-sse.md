# Task 11: Messages Endpoint with SSE Streaming

Send a chat message, call Anthropic, stream the response back via SSE.

## Steps

- [ ] `POST /api/papers/:id/chats/:chatId/messages` — accepts message request body
- [ ] Store the user message in DB
- [ ] Build conversation history from stored messages
- [ ] Call Anthropic client with history + optional selected/surrounding text
- [ ] Stream assistant response back to client as SSE events
- [ ] Store the completed assistant message in DB after stream finishes
- [ ] Handler tests: mock the Anthropic client, verify SSE output and DB writes
- [ ] Error handling: return appropriate errors if Anthropic call fails
