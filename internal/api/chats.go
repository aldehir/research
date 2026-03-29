package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/aldehir/research/internal/store"
)

func handleListChatSessions(db *sql.DB, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		paperID := r.PathValue("id")

		_, err := store.GetPaper(db, paperID)
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "paper not found", logger)
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to get paper", logger)
			return
		}

		sessions, err := store.ListChatSessions(db, paperID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to list chat sessions", logger)
			return
		}
		if sessions == nil {
			sessions = []store.ChatSession{}
		}
		writeJSON(w, http.StatusOK, sessions)
	}
}

func handleCreateChatSession(db *sql.DB, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		paperID := r.PathValue("id")

		_, err := store.GetPaper(db, paperID)
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "paper not found", logger)
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to get paper", logger)
			return
		}

		var body struct {
			Title string `json:"title"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON body", logger)
			return
		}

		id, err := newUUID()
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to generate ID", logger)
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
			writeError(w, http.StatusInternalServerError, "failed to create chat session", logger)
			return
		}

		writeJSON(w, http.StatusCreated, session)
	}
}

// messageWithAttachments extends a store.Message with its attachments for API responses.
type messageWithAttachments struct {
	store.Message
	Attachments []attachmentResponse `json:"attachments,omitempty"`
}

type attachmentResponse struct {
	ID   string `json:"id"`
	Text string `json:"text"`
	Page int    `json:"page"`
}

func handleGetChatSession(db *sql.DB, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		chatID := r.PathValue("chatId")

		result, err := store.GetChatSessionWithMessages(db, chatID)
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "chat session not found", logger)
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to get chat session", logger)
			return
		}

		// Load attachments for all messages in this chat
		chatAtts, err := store.ListAttachmentsByChat(db, chatID)
		if err != nil {
			logger.Warn("failed to load attachments", "chat_id", chatID, "error", err)
		}

		// Group attachments by message ID
		attsByMsg := make(map[string][]attachmentResponse)
		for _, a := range chatAtts {
			attsByMsg[a.MessageID] = append(attsByMsg[a.MessageID], attachmentResponse{
				ID:   a.ID,
				Text: a.Text,
				Page: a.Page,
			})
		}

		// Build enriched messages, filtering out tool interaction messages
		// (those have content_blocks but no user-visible content)
		var msgs []messageWithAttachments
		for _, m := range result.Messages {
			if m.ContentBlocks != nil {
				continue
			}
			msgs = append(msgs, messageWithAttachments{Message: m, Attachments: attsByMsg[m.ID]})
		}
		if msgs == nil {
			msgs = []messageWithAttachments{}
		}

		resp := struct {
			store.ChatSession
			Messages []messageWithAttachments `json:"messages"`
		}{
			ChatSession: result.ChatSession,
			Messages:    msgs,
		}

		writeJSON(w, http.StatusOK, resp)
	}
}

func handleDeleteChatSession(db *sql.DB, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		chatID := r.PathValue("chatId")

		err := store.DeleteChatSession(db, chatID)
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "chat session not found", logger)
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to delete chat session", logger)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
