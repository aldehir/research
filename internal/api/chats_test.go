package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aldehir/research/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func seedTestPaper(t *testing.T, tdb *store.TestDB) store.Paper {
	t.Helper()
	p := store.Paper{
		ID:        "paper-1",
		Title:     "Test Paper",
		FilePath:  "/papers/test.pdf",
		FileSize:  12345,
		CreatedAt: "2026-03-28T00:00:00Z",
	}
	require.NoError(t, store.CreatePaper(tdb.DB, p))
	return p
}

func TestListChatSessions(t *testing.T) {
	t.Run("empty returns empty array", func(t *testing.T) {
		mux, tdb, _ := testMux(t)
		seedTestPaper(t, tdb)

		req := httptest.NewRequest(http.MethodGet, "/api/papers/paper-1/chats", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

		var sessions []store.ChatSession
		err := json.NewDecoder(rec.Body).Decode(&sessions)
		require.NoError(t, err)
		assert.Empty(t, sessions)
	})

	t.Run("non-existent paper returns 404", func(t *testing.T) {
		mux, _, _ := testMux(t)

		req := httptest.NewRequest(http.MethodGet, "/api/papers/missing/chats", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})
}

func TestCreateChatSession(t *testing.T) {
	t.Run("creates chat session with default title", func(t *testing.T) {
		mux, tdb, _ := testMux(t)
		seedTestPaper(t, tdb)

		req := httptest.NewRequest(http.MethodPost, "/api/papers/paper-1/chats", strings.NewReader(`{}`))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusCreated, rec.Code)
		assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

		var session store.ChatSession
		err := json.NewDecoder(rec.Body).Decode(&session)
		require.NoError(t, err)
		assert.NotEmpty(t, session.ID)
		assert.Equal(t, "paper-1", session.PaperID)
		assert.NotEmpty(t, session.Title)
		assert.NotEmpty(t, session.CreatedAt)
	})

	t.Run("creates chat session with custom title", func(t *testing.T) {
		mux, tdb, _ := testMux(t)
		seedTestPaper(t, tdb)

		req := httptest.NewRequest(http.MethodPost, "/api/papers/paper-1/chats", strings.NewReader(`{"title":"My Chat"}`))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusCreated, rec.Code)

		var session store.ChatSession
		err := json.NewDecoder(rec.Body).Decode(&session)
		require.NoError(t, err)
		assert.Equal(t, "My Chat", session.Title)
	})

	t.Run("non-existent paper returns 404", func(t *testing.T) {
		mux, _, _ := testMux(t)

		req := httptest.NewRequest(http.MethodPost, "/api/papers/missing/chats", strings.NewReader(`{}`))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})
}

func TestGetChatSessionAPI(t *testing.T) {
	t.Run("returns session with messages", func(t *testing.T) {
		mux, tdb, _ := testMux(t)
		seedTestPaper(t, tdb)

		session := store.ChatSession{
			ID:        "chat-1",
			PaperID:   "paper-1",
			Title:     "Test Chat",
			CreatedAt: "2026-03-28T10:00:00Z",
		}
		require.NoError(t, store.CreateChatSession(tdb.DB, session))

		msg := store.Message{
			ID:            "msg-1",
			ChatSessionID: "chat-1",
			Role:          "user",
			Content:       "Hello",
			CreatedAt:     "2026-03-28T10:01:00Z",
		}
		require.NoError(t, store.CreateMessage(tdb.DB, msg))

		req := httptest.NewRequest(http.MethodGet, "/api/papers/paper-1/chats/chat-1", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var got store.ChatSessionWithMessages
		err := json.NewDecoder(rec.Body).Decode(&got)
		require.NoError(t, err)
		assert.Equal(t, "chat-1", got.ID)
		require.Len(t, got.Messages, 1)
		assert.Equal(t, "msg-1", got.Messages[0].ID)
	})

	t.Run("non-existent chat returns 404", func(t *testing.T) {
		mux, tdb, _ := testMux(t)
		seedTestPaper(t, tdb)

		req := httptest.NewRequest(http.MethodGet, "/api/papers/paper-1/chats/missing", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})
}

func TestDeleteChatSessionAPI(t *testing.T) {
	t.Run("deletes existing chat returns 204", func(t *testing.T) {
		mux, tdb, _ := testMux(t)
		seedTestPaper(t, tdb)

		session := store.ChatSession{
			ID:        "chat-1",
			PaperID:   "paper-1",
			Title:     "Test Chat",
			CreatedAt: "2026-03-28T10:00:00Z",
		}
		require.NoError(t, store.CreateChatSession(tdb.DB, session))

		req := httptest.NewRequest(http.MethodDelete, "/api/papers/paper-1/chats/chat-1", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNoContent, rec.Code)

		// Confirm it's gone
		req = httptest.NewRequest(http.MethodGet, "/api/papers/paper-1/chats/chat-1", nil)
		rec = httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("delete non-existent chat returns 404", func(t *testing.T) {
		mux, tdb, _ := testMux(t)
		seedTestPaper(t, tdb)

		req := httptest.NewRequest(http.MethodDelete, "/api/papers/paper-1/chats/missing", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})
}
