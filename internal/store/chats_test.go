package store

import (
	"database/sql"
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

	selectedText := "some selected text"
	surroundingText := "context around it"

	m1 := Message{
		ID:            "msg-1",
		ChatSessionID: "chat-1",
		Role:          "user",
		Content:       "Hello",
		SelectedText:  &selectedText,
		SurroundingText: &surroundingText,
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
	assert.Equal(t, &selectedText, got.Messages[0].SelectedText)
	assert.Nil(t, got.Messages[1].SelectedText)
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
