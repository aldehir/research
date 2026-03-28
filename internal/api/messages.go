package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/aldehir/research/internal/anthropic"
	"github.com/aldehir/research/internal/pdf"
	"github.com/aldehir/research/internal/store"
)

// ChatStreamer is an interface for streaming chat responses.
// The real *anthropic.Client satisfies this interface.
type ChatStreamer interface {
	Stream(ctx context.Context, req anthropic.Request) (<-chan anthropic.StreamEvent, error)
}

const maxToolLoopIterations = 10

func handleSendMessage(db *sql.DB, storage *pdf.Storage, chat ChatStreamer, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		paperID := r.PathValue("id")
		chatID := r.PathValue("chatId")

		// Check if chat streamer is available
		if chat == nil {
			writeError(w, http.StatusServiceUnavailable, "chat features unavailable", logger)
			return
		}

		// Parse request body
		var body struct {
			Content         string `json:"content"`
			SelectedText    string `json:"selected_text"`
			SurroundingText string `json:"surrounding_text"`
			CurrentPage     int    `json:"current_page"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON body", logger)
			return
		}

		if body.Content == "" {
			writeError(w, http.StatusBadRequest, "content is required", logger)
			return
		}

		// Validate chat session exists
		_, err := store.GetChatSession(db, chatID)
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "chat session not found", logger)
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to get chat session", logger)
			return
		}

		// Store user message
		msgID, err := newUUID()
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to generate ID", logger)
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
			writeError(w, http.StatusInternalServerError, "failed to store message", logger)
			return
		}

		// Load all messages for conversation history
		messages, err := store.ListMessages(db, chatID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to load messages", logger)
			return
		}

		// Convert to anthropic messages, appending viewer context to the latest user message
		var anthropicMessages []anthropic.Message
		for _, m := range messages {
			content := m.Content
			if m.Role == "user" && m.ID == userMsg.ID {
				content = appendViewerContext(content, body.CurrentPage, body.SelectedText)
			}
			anthropicMessages = append(anthropicMessages, anthropic.Message{
				Role:    m.Role,
				Content: content,
			})
		}

		// Look up paper metadata for prompt context
		var docTitle, docAuthor, docDate, pdfPath string
		var totalPages int
		if paper, err := store.GetPaper(db, paperID); err == nil {
			docTitle = paper.Title
			pdfPath = storage.Path(paperID)
			if paper.Author != nil {
				docAuthor = *paper.Author
			}
			if paper.PublishedDate != nil {
				docDate = *paper.PublishedDate
			}
			if paper.PageCount != nil {
				totalPages = *paper.PageCount
			}
		}

		// Build request
		req := anthropic.Request{
			Messages:       anthropicMessages,
			DocumentTitle:  docTitle,
			DocumentAuthor: docAuthor,
			DocumentDate:   docDate,
			TotalPages:     totalPages,
			Tools:          anthropic.PDFTools(),
		}

		// Set SSE headers before calling Stream
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		flusher, _ := w.(http.Flusher)
		flush := func() {
			if flusher != nil {
				flusher.Flush()
			}
		}

		sendSSE := func(v any) {
			fmt.Fprintf(w, "data: %s\n\n", mustJSON(v))
			flush()
		}

		var fullText strings.Builder
		responseStart := time.Now()
		var toolIterations int

		// Tool execution loop
		for i := 0; i < maxToolLoopIterations; i++ {
			ch, err := chat.Stream(r.Context(), req)
			if err != nil {
				logger.Error("stream start failed", "chat_id", chatID, "error", err)
				sendSSE(sseResponse{Type: "error", Error: err.Error()})
				return
			}

			var toolCalls []anthropic.StreamEvent
			for ev := range ch {
				switch ev.Type {
				case "content_block_delta":
					fullText.WriteString(ev.Text)
					sendSSE(sseResponse{Type: "delta", Text: ev.Text})
				case "tool_use":
					toolCalls = append(toolCalls, ev)
				case "message_stop":
					// Don't emit done yet if we have tool calls to process
				}
			}

			if len(toolCalls) == 0 {
				// No tool calls — we're done
				sendSSE(sseResponse{Type: "done"})
				break
			}

			toolIterations++
			logger.Info("tool_loop_iteration", "iteration", toolIterations, "tool_call_count", len(toolCalls))

			// Process tool calls
			// Build assistant message with tool_use blocks
			var assistantBlocks []anthropic.ContentBlock
			for _, tc := range toolCalls {
				logger.Debug("tool_call", "name", tc.ToolName, "args", tc.ToolInput)
				assistantBlocks = append(assistantBlocks, anthropic.ContentBlock{
					Type:  "tool_use",
					ID:    tc.ToolUseID,
					Name:  tc.ToolName,
					Input: json.RawMessage(tc.ToolInput),
				})

				// Send tool_call SSE to client for UI
				sendSSE(sseToolCall{
					Type: "tool_call",
					Name: tc.ToolName,
					Args: json.RawMessage(tc.ToolInput),
				})
			}

			// Append assistant tool_use message
			req.Messages = append(req.Messages, anthropic.Message{
				Role:          "assistant",
				ContentBlocks: assistantBlocks,
			})

			// Execute tools and build tool_result blocks
			tctx := toolContext{db: db, paperID: paperID, pdfPath: pdfPath}
			var resultBlocks []anthropic.ContentBlock
			for _, tc := range toolCalls {
				toolStart := time.Now()
				result := executeToolCall(tc.ToolName, tc.ToolInput, tctx, logger)
				logger.Info("tool_result",
					"name", tc.ToolName,
					"result_length", len(result),
					"duration", time.Since(toolStart),
				)
				resultBlocks = append(resultBlocks, anthropic.ContentBlock{
					Type:      "tool_result",
					ToolUseID: tc.ToolUseID,
					Content:   result,
				})
				sendSSE(sseToolResult{
					Type:    "tool_result",
					Name:    tc.ToolName,
					Text:    result,
					Preview: truncatePreview(result, toolResultPreviewLen),
				})
			}

			// Append user message with tool_result blocks
			req.Messages = append(req.Messages, anthropic.Message{
				Role:          "user",
				ContentBlocks: resultBlocks,
			})
		}

		logger.Info("response_complete",
			"response_length", fullText.Len(),
			"tool_iterations", toolIterations,
			"total_duration", time.Since(responseStart),
		)

		// Store assistant message (final text only)
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

type sseToolCall struct {
	Type string          `json:"type"`
	Name string          `json:"name"`
	Args json.RawMessage `json:"args"`
}

type sseToolResult struct {
	Type    string `json:"type"`
	Name    string `json:"name"`
	Text    string `json:"text"`
	Preview string `json:"preview"`
}

const toolResultPreviewLen = 200

func truncatePreview(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func executeToolCall(name, input string, tc toolContext, logger *slog.Logger) string {
	switch name {
	case "search_pdf":
		var args struct {
			Query string `json:"query"`
		}
		if err := json.Unmarshal([]byte(input), &args); err != nil {
			return fmt.Sprintf("Error parsing arguments: %v", err)
		}

		// Try FTS5 index first
		if tc.db != nil && tc.paperID != "" {
			results, err := store.SearchPageText(tc.db, tc.paperID, args.Query)
			if err == nil && len(results) > 0 {
				b, _ := json.Marshal(results)
				return string(b)
			}
		}

		// Fall back to pdftotext
		results, err := pdf.SearchText(tc.pdfPath, args.Query)
		if err != nil {
			logger.Warn("search_pdf failed", "error", err)
			return fmt.Sprintf("Error searching PDF: %v", err)
		}
		if len(results) == 0 {
			return "No matches found."
		}
		b, _ := json.Marshal(results)
		return string(b)

	case "read_page":
		var args struct {
			Page int `json:"page"`
		}
		if err := json.Unmarshal([]byte(input), &args); err != nil {
			return fmt.Sprintf("Error parsing arguments: %v", err)
		}

		// Try indexed text first
		if tc.db != nil && tc.paperID != "" {
			text, err := store.GetPageText(tc.db, tc.paperID, args.Page)
			if err == nil {
				return text
			}
		}

		// Fall back to pdftotext
		text, err := pdf.ExtractPageText(tc.pdfPath, args.Page)
		if err != nil {
			logger.Warn("read_page failed", "error", err)
			return fmt.Sprintf("Error reading page: %v", err)
		}
		return text

	case "go_to_page":
		// Client-side tool — return success, the SSE event was already sent
		var args struct {
			Page int `json:"page"`
		}
		if err := json.Unmarshal([]byte(input), &args); err != nil {
			return fmt.Sprintf("Error parsing arguments: %v", err)
		}
		return fmt.Sprintf("Navigated to page %d.", args.Page)

	default:
		return fmt.Sprintf("Unknown tool: %s", name)
	}
}

// toolContext holds context needed for tool execution.
type toolContext struct {
	db      *sql.DB
	paperID string
	pdfPath string
}

func appendViewerContext(content string, currentPage int, selectedText string) string {
	if currentPage == 0 && selectedText == "" {
		return content
	}

	var b strings.Builder
	b.WriteString(content)
	b.WriteString("\n\n[Viewer context]")
	if currentPage > 0 {
		b.WriteString(fmt.Sprintf("\nCurrent page: %d", currentPage))
	}
	if selectedText != "" {
		b.WriteString(fmt.Sprintf("\nSelected text: %s", selectedText))
	}
	return b.String()
}

func mustJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}
