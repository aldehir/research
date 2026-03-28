package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/aldehir/research/internal/store"
)

func handleListChatSessions(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		paperID := r.PathValue("id")

		_, err := store.GetPaper(db, paperID)
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "paper not found")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to get paper")
			return
		}

		sessions, err := store.ListChatSessions(db, paperID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to list chat sessions")
			return
		}
		if sessions == nil {
			sessions = []store.ChatSession{}
		}
		writeJSON(w, http.StatusOK, sessions)
	}
}

func handleCreateChatSession(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		paperID := r.PathValue("id")

		_, err := store.GetPaper(db, paperID)
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "paper not found")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to get paper")
			return
		}

		var body struct {
			Title string `json:"title"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON body")
			return
		}

		id, err := newUUID()
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to generate ID")
			return
		}

		now := time.Now().UTC()
		title := body.Title
		if title == "" {
			title = "Chat " + now.Format("2006-01-02 15:04:05")
		}

		session := store.ChatSession{
			ID:        id,
			PaperID:   paperID,
			Title:     title,
			CreatedAt: now.Format(time.RFC3339),
		}

		if err := store.CreateChatSession(db, session); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to create chat session")
			return
		}

		writeJSON(w, http.StatusCreated, session)
	}
}

func handleGetChatSession(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		chatID := r.PathValue("chatId")

		result, err := store.GetChatSessionWithMessages(db, chatID)
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "chat session not found")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to get chat session")
			return
		}

		writeJSON(w, http.StatusOK, result)
	}
}

func handleDeleteChatSession(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		chatID := r.PathValue("chatId")

		err := store.DeleteChatSession(db, chatID)
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "chat session not found")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to delete chat session")
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
