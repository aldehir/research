# Task 09: Chat Sessions API

CRUD endpoints for chat sessions tied to papers.

## Store Layer

- [ ] `internal/store/chats.go` — Create, List (by paper), GetByID (with messages), Delete
- [ ] Table-driven tests for each method
- [ ] Deleting a chat session cascades to its messages

## API Layer

- [ ] `GET /api/papers/:id/chats` — list sessions for a paper
- [ ] `POST /api/papers/:id/chats` — create a session (auto-generate title)
- [ ] `GET /api/papers/:id/chats/:chatId` — get session with full message history
- [ ] `DELETE /api/papers/:id/chats/:chatId` — delete session and messages
- [ ] Handler tests
- [ ] Validate paper exists before creating a chat session
