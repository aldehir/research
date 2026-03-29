package store

import (
	"database/sql"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func seedPaper(t *testing.T, db *sql.DB) Paper {
	t.Helper()
	p := Paper{
		ID:        "paper-1",
		Title:     "Test Paper",
		FilePath:  "/papers/test.pdf",
		FileSize:  12345,
		CreatedAt: "2026-03-28T00:00:00Z",
	}
	require.NoError(t, CreatePaper(db, p))
	return p
}

func TestCreateAndGetChatSession(t *testing.T) {
	db := testDB(t)
	paper := seedPaper(t, db)

	session := ChatSession{
		ID:        "chat-1",
		PaperID:   paper.ID,
		Title:     "Chat 2026-03-28",
		CreatedAt: "2026-03-28T10:00:00Z",
	}

	err := CreateChatSession(db, session)
	require.NoError(t, err)

	got, err := GetChatSession(db, "chat-1")
	require.NoError(t, err)
	assert.Equal(t, session, got)
}

func TestListChatSessions(t *testing.T) {
	db := testDB(t)
	paper := seedPaper(t, db)

	t.Run("empty returns empty slice", func(t *testing.T) {
		sessions, err := ListChatSessions(db, paper.ID)
		require.NoError(t, err)
		assert.Empty(t, sessions)
	})

	t.Run("returns sessions ordered by created_at desc", func(t *testing.T) {
		s1 := ChatSession{ID: "chat-1", PaperID: paper.ID, Title: "Older", CreatedAt: "2026-03-01T00:00:00Z"}
		s2 := ChatSession{ID: "chat-2", PaperID: paper.ID, Title: "Newer", CreatedAt: "2026-03-28T00:00:00Z"}
		require.NoError(t, CreateChatSession(db, s1))
		require.NoError(t, CreateChatSession(db, s2))

		sessions, err := ListChatSessions(db, paper.ID)
		require.NoError(t, err)
		require.Len(t, sessions, 2)
		assert.Equal(t, "chat-2", sessions[0].ID)
		assert.Equal(t, "chat-1", sessions[1].ID)
	})
}

func TestGetChatSessionNotFound(t *testing.T) {
	db := testDB(t)

	_, err := GetChatSession(db, "nonexistent")
	assert.ErrorIs(t, err, sql.ErrNoRows)
}

func TestGetChatSessionWithMessages(t *testing.T) {
	db := testDB(t)
	paper := seedPaper(t, db)

	session := ChatSession{
		ID:        "chat-1",
		PaperID:   paper.ID,
		Title:     "Test Chat",
		CreatedAt: "2026-03-28T10:00:00Z",
	}
	require.NoError(t, CreateChatSession(db, session))

	m1 := Message{
		ID:            "msg-1",
		ChatSessionID: "chat-1",
		Role:          "user",
		Content:       "Hello",
		CreatedAt:     "2026-03-28T10:01:00Z",
	}
	m2 := Message{
		ID:            "msg-2",
		ChatSessionID: "chat-1",
		Role:          "assistant",
		Content:       "Hi there",
		CreatedAt:     "2026-03-28T10:02:00Z",
	}
	require.NoError(t, CreateMessage(db, m1))
	require.NoError(t, CreateMessage(db, m2))

	got, err := GetChatSessionWithMessages(db, "chat-1")
	require.NoError(t, err)
	assert.Equal(t, session, got.ChatSession)
	require.Len(t, got.Messages, 2)
	assert.Equal(t, "msg-1", got.Messages[0].ID)
	assert.Equal(t, "msg-2", got.Messages[1].ID)
}

func TestGetChatSessionWithMessagesNotFound(t *testing.T) {
	db := testDB(t)

	_, err := GetChatSessionWithMessages(db, "nonexistent")
	assert.ErrorIs(t, err, sql.ErrNoRows)
}

func TestDeleteChatSession(t *testing.T) {
	db := testDB(t)
	paper := seedPaper(t, db)

	session := ChatSession{
		ID:        "chat-1",
		PaperID:   paper.ID,
		Title:     "To Delete",
		CreatedAt: "2026-03-28T10:00:00Z",
	}
	require.NoError(t, CreateChatSession(db, session))

	msg := Message{
		ID:            "msg-1",
		ChatSessionID: "chat-1",
		Role:          "user",
		Content:       "Hello",
		CreatedAt:     "2026-03-28T10:01:00Z",
	}
	require.NoError(t, CreateMessage(db, msg))

	err := DeleteChatSession(db, "chat-1")
	require.NoError(t, err)

	_, err = GetChatSession(db, "chat-1")
	assert.ErrorIs(t, err, sql.ErrNoRows)

	// Messages should be cascade deleted
	messages, err := ListMessages(db, "chat-1")
	require.NoError(t, err)
	assert.Empty(t, messages)
}

func TestDeleteChatSessionNotFound(t *testing.T) {
	db := testDB(t)

	err := DeleteChatSession(db, "nonexistent")
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestCreateAndListMessages(t *testing.T) {
	db := testDB(t)
	paper := seedPaper(t, db)

	session := ChatSession{
		ID:        "chat-1",
		PaperID:   paper.ID,
		Title:     "Test Chat",
		CreatedAt: "2026-03-28T10:00:00Z",
	}
	require.NoError(t, CreateChatSession(db, session))

	m1 := Message{
		ID:            "msg-1",
		ChatSessionID: "chat-1",
		Role:          "user",
		Content:       "First",
		CreatedAt:     "2026-03-28T10:01:00Z",
	}
	m2 := Message{
		ID:            "msg-2",
		ChatSessionID: "chat-1",
		Role:          "assistant",
		Content:       "Second",
		CreatedAt:     "2026-03-28T10:02:00Z",
	}
	require.NoError(t, CreateMessage(db, m1))
	require.NoError(t, CreateMessage(db, m2))

	messages, err := ListMessages(db, "chat-1")
	require.NoError(t, err)
	require.Len(t, messages, 2)
	assert.Equal(t, "msg-1", messages[0].ID)
	assert.Equal(t, "msg-2", messages[1].ID)
	assert.Equal(t, "user", messages[0].Role)
	assert.Equal(t, "assistant", messages[1].Role)
}

func TestCreateAndListMessages_WithContentBlocks(t *testing.T) {
	db := testDB(t)
	paper := seedPaper(t, db)

	session := ChatSession{
		ID:        "chat-1",
		PaperID:   paper.ID,
		Title:     "Test Chat",
		CreatedAt: "2026-03-28T10:00:00Z",
	}
	require.NoError(t, CreateChatSession(db, session))

	blocks := `[{"type":"tool_use","id":"toolu_1","name":"read_page","input":{"page":3}}]`

	m := Message{
		ID:            "msg-1",
		ChatSessionID: "chat-1",
		Role:          "assistant",
		Content:       "",
		ContentBlocks: &blocks,
		CreatedAt:     "2026-03-28T10:01:00Z",
	}
	require.NoError(t, CreateMessage(db, m))

	messages, err := ListMessages(db, "chat-1")
	require.NoError(t, err)
	require.Len(t, messages, 1)
	assert.Equal(t, "msg-1", messages[0].ID)
	require.NotNil(t, messages[0].ContentBlocks, "content_blocks should be loaded")
	assert.Equal(t, blocks, *messages[0].ContentBlocks)

	// Verify valid JSON
	var parsed []json.RawMessage
	require.NoError(t, json.Unmarshal([]byte(*messages[0].ContentBlocks), &parsed))
	assert.Len(t, parsed, 1)
}

func TestCreateAttachmentAndListByMessage(t *testing.T) {
	db := testDB(t)
	paper := seedPaper(t, db)

	session := ChatSession{
		ID:        "chat-1",
		PaperID:   paper.ID,
		Title:     "Test Chat",
		CreatedAt: "2026-03-28T10:00:00Z",
	}
	require.NoError(t, CreateChatSession(db, session))

	msg := Message{
		ID:            "msg-1",
		ChatSessionID: "chat-1",
		Role:          "user",
		Content:       "Hello",
		CreatedAt:     "2026-03-28T10:01:00Z",
	}
	require.NoError(t, CreateMessage(db, msg))

	att := Attachment{
		ID:        "att-1",
		MessageID: "msg-1",
		FilePath:  "/data/attachments/att-1.png",
		Text:      "Figure 1: Results",
		Page:      3,
		CreatedAt: "2026-03-28T10:01:00Z",
	}
	require.NoError(t, CreateAttachment(db, att))

	atts, err := ListAttachmentsByMessage(db, "msg-1")
	require.NoError(t, err)
	require.Len(t, atts, 1)
	assert.Equal(t, "att-1", atts[0].ID)
	assert.Equal(t, "msg-1", atts[0].MessageID)
	assert.Equal(t, "/data/attachments/att-1.png", atts[0].FilePath)
	assert.Equal(t, "Figure 1: Results", atts[0].Text)
	assert.Equal(t, 3, atts[0].Page)
}

func TestListAttachmentsByChat(t *testing.T) {
	db := testDB(t)
	paper := seedPaper(t, db)

	session := ChatSession{
		ID:        "chat-1",
		PaperID:   paper.ID,
		Title:     "Test Chat",
		CreatedAt: "2026-03-28T10:00:00Z",
	}
	require.NoError(t, CreateChatSession(db, session))

	m1 := Message{ID: "msg-1", ChatSessionID: "chat-1", Role: "user", Content: "First", CreatedAt: "2026-03-28T10:01:00Z"}
	m2 := Message{ID: "msg-2", ChatSessionID: "chat-1", Role: "user", Content: "Second", CreatedAt: "2026-03-28T10:02:00Z"}
	require.NoError(t, CreateMessage(db, m1))
	require.NoError(t, CreateMessage(db, m2))

	a1 := Attachment{ID: "att-1", MessageID: "msg-1", FilePath: "/att-1.png", Text: "Fig 1", Page: 1, CreatedAt: "2026-03-28T10:01:00Z"}
	a2 := Attachment{ID: "att-2", MessageID: "msg-2", FilePath: "/att-2.png", Text: "Fig 2", Page: 2, CreatedAt: "2026-03-28T10:02:00Z"}
	require.NoError(t, CreateAttachment(db, a1))
	require.NoError(t, CreateAttachment(db, a2))

	atts, err := ListAttachmentsByChat(db, "chat-1")
	require.NoError(t, err)
	require.Len(t, atts, 2)
	assert.Equal(t, "att-1", atts[0].ID)
	assert.Equal(t, "att-2", atts[1].ID)
}

func TestAttachmentCascadeDeleteOnMessage(t *testing.T) {
	db := testDB(t)
	paper := seedPaper(t, db)

	session := ChatSession{ID: "chat-1", PaperID: paper.ID, Title: "Test", CreatedAt: "2026-03-28T10:00:00Z"}
	require.NoError(t, CreateChatSession(db, session))

	msg := Message{ID: "msg-1", ChatSessionID: "chat-1", Role: "user", Content: "Hello", CreatedAt: "2026-03-28T10:01:00Z"}
	require.NoError(t, CreateMessage(db, msg))

	att := Attachment{ID: "att-1", MessageID: "msg-1", FilePath: "/att-1.png", Text: "text", Page: 1, CreatedAt: "2026-03-28T10:01:00Z"}
	require.NoError(t, CreateAttachment(db, att))

	// Delete the session (cascades to messages -> attachments)
	require.NoError(t, DeleteChatSession(db, "chat-1"))

	atts, err := ListAttachmentsByMessage(db, "msg-1")
	require.NoError(t, err)
	assert.Empty(t, atts)
}

func TestCreateAndListMessages_NilContentBlocks(t *testing.T) {
	db := testDB(t)
	paper := seedPaper(t, db)

	session := ChatSession{
		ID:        "chat-1",
		PaperID:   paper.ID,
		Title:     "Test Chat",
		CreatedAt: "2026-03-28T10:00:00Z",
	}
	require.NoError(t, CreateChatSession(db, session))

	m := Message{
		ID:            "msg-1",
		ChatSessionID: "chat-1",
		Role:          "user",
		Content:       "Hello",
		CreatedAt:     "2026-03-28T10:01:00Z",
	}
	require.NoError(t, CreateMessage(db, m))

	messages, err := ListMessages(db, "chat-1")
	require.NoError(t, err)
	require.Len(t, messages, 1)
	assert.Nil(t, messages[0].ContentBlocks, "content_blocks should be nil for plain text messages")
	assert.Equal(t, "Hello", messages[0].Content)
}
