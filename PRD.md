# PRD: Research Paper Reader

## Overview

A personal web application for reading research papers (PDFs) with integrated
LLM chat. Upload papers, read them in the browser, select text, and ask Claude
questions about what you're reading.

## Goals

- Render PDFs with full text selection support
- Chat with Claude about selected passages, using surrounding context for
  better answers
- Store PDFs server-side for access from any device
- Single-user tool — no authentication

## Architecture

| Layer    | Stack                                      |
|----------|--------------------------------------------|
| Frontend | TypeScript, SvelteKit 5 (SPA mode), Vite   |
| Backend  | Go (HTTP API)                               |
| PDF      | pdf.js (Mozilla)                            |
| LLM      | Anthropic Messages API (Claude)             |
| Storage  | Filesystem (PDFs), SQLite (metadata + chat) |

### Why SPA mode?

The constraint is no server-side rendering. SvelteKit will be configured with
`adapter-static` and all routing/rendering happens client-side.

### Why SQLite?

Single-user, self-hosted tool. SQLite keeps deployment simple (no external
database) while still giving us structured queries for metadata and chat
history.

## Features

### 1. PDF Library

Users manage a personal library of uploaded research papers.

- **Upload**: Drag-and-drop zone + file picker button. Accept `.pdf` only.
- **List view**: Shows all uploaded papers with title, upload date, and file
  size.
- **Delete**: Remove a paper and its associated chat sessions.
- **Storage**: PDFs stored on the server filesystem. Metadata (title, upload
  timestamp, file size) stored in SQLite.

### 2. PDF Viewer

Full in-browser PDF rendering using Mozilla's pdf.js.

- **Rendering**: Page-by-page rendering with scroll-based navigation.
- **Text selection**: Native text layer from pdf.js for selecting passages.
- **Zoom**: Zoom in/out controls.
- **Page navigation**: Jump to page, previous/next page.

### 3. LLM Chat

Context-aware chat sessions tied to individual PDFs.

- **Scope**: Each chat session is associated with exactly one PDF.
- **Selection context**: When the user selects text and initiates a chat
  message, the system includes:
  1. The selected text (highlighted/quoted in the prompt)
  2. Surrounding text from the same page for context
- **Free-form chat**: Users can also type questions without a selection.
- **Chat history**: Conversations are stored in SQLite and persist across
  sessions.
- **Multiple sessions**: A user can have multiple chat sessions per PDF.
- **Streaming**: Responses stream back to the UI token-by-token via
  server-sent events (SSE).

### 4. API Key Configuration

The Anthropic API key is provided via the `ANTHROPIC_API_KEY` environment
variable on the server. The server validates the key is set at startup and
returns a clear error if missing.

## API Design

All endpoints are prefixed with `/api`.

### Papers

| Method | Path               | Description              |
|--------|--------------------|--------------------------|
| GET    | /api/papers        | List all papers          |
| POST   | /api/papers        | Upload a PDF             |
| GET    | /api/papers/:id    | Get paper metadata       |
| DELETE | /api/papers/:id    | Delete paper + chats     |
| GET    | /api/papers/:id/pdf| Serve the PDF file       |

### Chat Sessions

| Method | Path                              | Description              |
|--------|-----------------------------------|--------------------------|
| GET    | /api/papers/:id/chats             | List chat sessions       |
| POST   | /api/papers/:id/chats             | Create a chat session    |
| GET    | /api/papers/:id/chats/:chatId     | Get chat with messages   |
| DELETE | /api/papers/:id/chats/:chatId     | Delete a chat session    |
| POST   | /api/papers/:id/chats/:chatId/messages | Send a message (SSE) |

### Message Request Body

```json
{
  "content": "What does this equation mean?",
  "selected_text": "E = mc^2",
  "surrounding_text": "...the famous equation E = mc^2 demonstrates..."
}
```

The `selected_text` and `surrounding_text` fields are optional. When present,
they are injected into the system prompt to give Claude focused context.

## Data Model

### papers

| Column     | Type    | Description                |
|------------|---------|----------------------------|
| id         | TEXT    | UUID primary key           |
| title      | TEXT    | Filename or extracted title |
| file_path  | TEXT    | Path to PDF on disk        |
| file_size  | INTEGER | Size in bytes              |
| created_at | TEXT    | ISO 8601 timestamp         |

### chat_sessions

| Column     | Type | Description                    |
|------------|------|--------------------------------|
| id         | TEXT | UUID primary key               |
| paper_id   | TEXT | FK to papers                   |
| title      | TEXT | Session title (auto-generated) |
| created_at | TEXT | ISO 8601 timestamp             |

### messages

| Column          | Type    | Description                          |
|-----------------|---------|--------------------------------------|
| id              | TEXT    | UUID primary key                     |
| chat_session_id | TEXT    | FK to chat_sessions                  |
| role            | TEXT    | "user" or "assistant"                |
| content         | TEXT    | Message text                         |
| selected_text   | TEXT    | Selected PDF text (nullable)         |
| surrounding_text| TEXT    | Context around selection (nullable)  |
| created_at      | TEXT    | ISO 8601 timestamp                   |

## UI Layout

```
+------------------------------------------+
|  Research Reader              [Upload]   |
+----------+-------------------------------+
|          |                    |           |
|  Paper   |    PDF Viewer      |   Chat   |
|  List    |                    |   Panel  |
|          |                    |           |
|  paper1  |  [rendered PDF]    | [session]|
|  paper2  |                    | [msgs]   |
|  paper3  |  [text selectable] |           |
|          |                    | [input]  |
+----------+-------------------------------+
```

Three-panel layout:
1. **Left sidebar**: Paper library list
2. **Center**: PDF viewer
3. **Right panel**: Chat (collapsible)

## Development Approach

**TDD (Red/Green/Refactor)** for all code:

- **Backend**: Table-driven Go tests. Test HTTP handlers, service logic, and
  database operations.
- **Frontend**: Vitest for unit/component tests. Test components, stores, and
  API client logic.

## Non-Goals (for v1)

- Multi-user / authentication
- PDF annotation or highlighting persistence
- Full-text search across papers
- Citation extraction or reference linking
- OCR for scanned PDFs
- Mobile-optimized layout
