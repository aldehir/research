# Task 12: Frontend — Chat Panel

Build the chat UI in the right panel.

## API Client

- [ ] Add chat session and message endpoints to the API client
- [ ] SSE client: consume `text/event-stream` for streaming responses
- [ ] Tests for API client additions

## Components

- [ ] Chat session list — shows sessions for the current paper, create new
- [ ] Message thread — renders conversation (user + assistant messages)
- [ ] Message input — text area with send button
- [ ] Streaming display — tokens appear as they arrive
- [ ] Collapsible panel (toggle visibility)

## State

- [ ] Active chat session state
- [ ] Messages state with streaming support
- [ ] Wire selected text from Task 08 into the message request
