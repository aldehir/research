package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/aldehir/research/internal/anthropic"
	"github.com/aldehir/research/internal/store"
)

// ChatStreamer is an interface for streaming chat responses.
// The real *anthropic.Client satisfies this interface.
type ChatStreamer interface {
	Stream(ctx context.Context, req anthropic.Request) (<-chan anthropic.StreamEvent, error)
}

func handleSendMessage(db *sql.DB, chat ChatStreamer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		chatID := r.PathValue("chatId")

		// Check if chat streamer is available
		if chat == nil {
			writeError(w, http.StatusServiceUnavailable, "chat features unavailable")
			return
		}

		// Parse request body
		var body struct {
			Content         string `json:"content"`
			SelectedText    string `json:"selected_text"`
			SurroundingText string `json:"surrounding_text"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON body")
			return
		}

		if body.Content == "" {
			writeError(w, http.StatusBadRequest, "content is required")
			return
		}

		// Validate chat session exists
		_, err := store.GetChatSession(db, chatID)
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "chat session not found")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to get chat session")
			return
		}

		// Store user message
		msgID, err := newUUID()
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to generate ID")
			return
		}

		var selectedText *string
		if body.SelectedText != "" {
			selectedText = &body.SelectedText
		}
		var surroundingText *string
		if body.SurroundingText != "" {
			surroundingText = &body.SurroundingText
		}

		userMsg := store.Message{
			ID:              msgID,
			ChatSessionID:   chatID,
			Role:            "user",
			Content:         body.Content,
			SelectedText:    selectedText,
			SurroundingText: surroundingText,
			CreatedAt:       time.Now().UTC().Format(time.RFC3339),
		}
		if err := store.CreateMessage(db, userMsg); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to store message")
			return
		}

		// Load all messages for conversation history
		messages, err := store.ListMessages(db, chatID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to load messages")
			return
		}

		// Convert to anthropic messages
		var anthropicMessages []anthropic.Message
		for _, m := range messages {
			anthropicMessages = append(anthropicMessages, anthropic.Message{
				Role:    m.Role,
				Content: m.Content,
			})
		}

		// Build request
		req := anthropic.Request{
			Messages:        anthropicMessages,
			SelectedText:    body.SelectedText,
			SurroundingText: body.SurroundingText,
		}

		// Set SSE headers before calling Stream
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		// Call stream
		ch, err := chat.Stream(r.Context(), req)
		if err != nil {
			// Already set SSE headers, so send error as SSE event
			fmt.Fprintf(w, "data: %s\n\n", mustJSON(sseResponse{Type: "error", Error: err.Error()}))
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
			return
		}

		// Stream events
		flusher, _ := w.(http.Flusher)
		var fullText strings.Builder

		for ev := range ch {
			switch ev.Type {
			case "content_block_delta":
				fullText.WriteString(ev.Text)
				fmt.Fprintf(w, "data: %s\n\n", mustJSON(sseResponse{Type: "delta", Text: ev.Text}))
			case "message_stop":
				fmt.Fprintf(w, "data: %s\n\n", mustJSON(sseResponse{Type: "done"}))
			}
			if flusher != nil {
				flusher.Flush()
			}
		}

		// Store assistant message
		assistantID, err := newUUID()
		if err != nil {
			return
		}
		assistantMsg := store.Message{
			ID:            assistantID,
			ChatSessionID: chatID,
			Role:          "assistant",
			Content:       fullText.String(),
			CreatedAt:     time.Now().UTC().Format(time.RFC3339),
		}
		store.CreateMessage(db, assistantMsg)
	}
}

type sseResponse struct {
	Type  string `json:"type"`
	Text  string `json:"text,omitempty"`
	Error string `json:"error,omitempty"`
}

func mustJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}
