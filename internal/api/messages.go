package api

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aldehir/research/internal/chat"
	"github.com/aldehir/research/internal/pdf"
	"github.com/aldehir/research/internal/store"
)


type attachment struct {
	ImageData string `json:"image_data"`
	Text      string `json:"text"`
	Page      int    `json:"page"`
}

const maxToolLoopIterations = 10

func handleSendMessage(db *sql.DB, storage *pdf.Storage, provider chat.Provider, dataDir string, registry *StreamRegistry, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		paperID := r.PathValue("id")
		chatID := r.PathValue("chatId")

		// Check if provider is available
		if provider == nil {
			writeError(w, http.StatusServiceUnavailable, "chat features unavailable", logger)
			return
		}

		// Reject if there's already a running stream for this chat.
		// Completed streams linger in the registry for reconnect but
		// should not block new messages — remove them first.
		if existing := registry.Get(chatID); existing != nil {
			if existing.Status() == StreamRunning {
				writeError(w, http.StatusConflict, "stream already in progress", logger)
				return
			}
			registry.Remove(chatID)
		}

		// Parse request body
		var body struct {
			Content     string       `json:"content"`
			CurrentPage int          `json:"current_page"`
			Attachments []attachment `json:"attachments"`
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

		userMsg := store.Message{
			ID:            msgID,
			ChatSessionID: chatID,
			Role:          "user",
			Content:       body.Content,
			CreatedAt:     time.Now().UTC().Format(time.RFC3339),
		}
		if err := store.CreateMessage(db, userMsg); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to store message", logger)
			return
		}

		// Save attachment images to disk and persist metadata
		if dataDir != "" && len(body.Attachments) > 0 {
			attDir := filepath.Join(dataDir, "attachments")
			if err := os.MkdirAll(attDir, 0o755); err != nil {
				logger.Error("failed to create attachments directory", "error", err)
			} else {
				for _, att := range body.Attachments {
					if att.ImageData == "" {
						continue
					}
					attID, err := newUUID()
					if err != nil {
						logger.Error("failed to generate attachment ID", "error", err)
						continue
					}
					imgBytes, err := base64.StdEncoding.DecodeString(att.ImageData)
					if err != nil {
						logger.Error("failed to decode attachment image", "error", err)
						continue
					}
					imgPath := filepath.Join(attDir, attID+".png")
					if err := os.WriteFile(imgPath, imgBytes, 0o644); err != nil {
						logger.Error("failed to write attachment image", "path", imgPath, "error", err)
						continue
					}
					storeAtt := store.Attachment{
						ID:        attID,
						MessageID: msgID,
						FilePath:  imgPath,
						Text:      att.Text,
						Page:      att.Page,
						CreatedAt: time.Now().UTC().Format(time.RFC3339),
					}
					if err := store.CreateAttachment(db, storeAtt); err != nil {
						logger.Error("failed to persist attachment", "id", attID, "error", err)
					} else {
						logger.Info("attachment saved", "id", attID, "message_id", msgID, "page", att.Page)
					}
				}
			}
		}

		// Load all messages for conversation history
		messages, err := store.ListMessages(db, chatID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to load messages", logger)
			return
		}

		// Load persisted attachments for the chat, grouped by message ID
		chatAtts, _ := store.ListAttachmentsByChat(db, chatID)
		attsByMsg := make(map[string][]store.Attachment)
		for _, a := range chatAtts {
			attsByMsg[a.MessageID] = append(attsByMsg[a.MessageID], a)
		}

		// Convert to domain messages, appending viewer context to the latest user message
		var chatMessages []chat.Message
		for _, m := range messages {
			// Messages with content_blocks: reconstruct structured content
			if m.ContentBlocks != nil {
				var parts []chat.Part
				if err := json.Unmarshal([]byte(*m.ContentBlocks), &parts); err == nil {
					chatMessages = append(chatMessages, chat.Message{
						Role:  chat.Role(m.Role),
						Parts: parts,
					})
					continue
				}
			}

			content := m.Content
			if m.Role == "user" && m.ID == userMsg.ID {
				content = appendViewerContext(content, body.CurrentPage)

				// Build multimodal message if attachments are present (current turn)
				if len(body.Attachments) > 0 {
					chatMessages = append(chatMessages, buildMultimodalUserMessage(content, body.Attachments))
					continue
				}
			}

			// Reconstruct multimodal message from persisted attachments (past turns)
			if m.Role == "user" && m.ID != userMsg.ID {
				if persistedAtts, ok := attsByMsg[m.ID]; ok && len(persistedAtts) > 0 {
					parts, err := buildPartsFromPersistedAttachments(content, persistedAtts, logger)
					if err == nil {
						chatMessages = append(chatMessages, chat.Message{
							Role:  chat.RoleUser,
							Parts: parts,
						})
						continue
					}
				}
			}

			chatMessages = append(chatMessages, chat.Message{
				Role:  chat.Role(m.Role),
				Parts: []chat.Part{{Kind: chat.PartText, Text: content}},
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
		req := chat.Request{
			SystemPrompt: chat.BuildSystemPrompt(chat.PromptContext{
				DocumentTitle:  docTitle,
				DocumentAuthor: docAuthor,
				DocumentDate:   docDate,
				TotalPages:     totalPages,
			}),
			Messages: chatMessages,
			Tools:    chat.PDFTools(),
		}

		// Start background stream
		stream, bgCtx := registry.Start(chatID)

		// Set SSE headers
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		flusher, _ := w.(http.Flusher)

		// Launch tool loop in a background goroutine
		go runToolLoop(bgCtx, stream, registry, db, provider, req, chatID, paperID, pdfPath, logger)

		// Relay events from the background stream to this HTTP response
		relayEvents(r.Context(), stream, w, flusher)
	}
}

// runToolLoop executes the LLM streaming and tool execution loop in the
// background. It writes SSE events to the ActiveStream buffer and persists
// the final assistant message to the database.
func runToolLoop(
	ctx context.Context,
	stream *ActiveStream,
	registry *StreamRegistry,
	db *sql.DB,
	provider chat.Provider,
	req chat.Request,
	chatID, paperID, pdfPath string,
	logger *slog.Logger,
) {
	defer func() {
		stream.SetStatus(StreamDone)
		// Schedule cleanup after a retention period
		ttl := registry.RetentionTTL
		go func() {
			time.Sleep(ttl)
			registry.Remove(chatID)
		}()
	}()

	emit := func(v any) {
		registry.Append(stream, SSEEvent{Data: mustJSON(v)})
	}

	var fullText strings.Builder
	responseStart := time.Now()
	var toolIterations int

	// Tool execution loop
	for i := 0; i < maxToolLoopIterations; i++ {
		ch, err := provider.Stream(ctx, req)
		if err != nil {
			logger.Error("stream start failed", "chat_id", chatID, "error", err)
			emit(sseResponse{Type: "error", Error: err.Error()})
			stream.SetStatus(StreamError)
			return
		}

		var toolCalls []chat.ToolCall
		var iterationText strings.Builder
		for ev := range ch {
			switch ev.Kind {
			case chat.EventDelta:
				fullText.WriteString(ev.Text)
				iterationText.WriteString(ev.Text)
				emit(sseResponse{Type: "delta", Text: ev.Text})
			case chat.EventToolCall:
				if ev.ToolCall != nil {
					toolCalls = append(toolCalls, *ev.ToolCall)
				}
			case chat.EventDone:
				// Don't emit done yet if we have tool calls to process
			}
		}

		if len(toolCalls) == 0 {
			// No tool calls — we're done
			emit(sseResponse{Type: "done"})
			break
		}

		toolIterations++
		logger.Info("tool_loop_iteration", "iteration", toolIterations, "tool_call_count", len(toolCalls))

		// Process tool calls
		// Build assistant message with tool call parts (include preceding text if any)
		var assistantParts []chat.Part
		if iterationText.Len() > 0 {
			assistantParts = append(assistantParts, chat.Part{
				Kind: chat.PartText,
				Text: iterationText.String(),
			})
		}
		for _, tc := range toolCalls {
			logger.Debug("tool_call", "name", tc.Name, "args", string(tc.Input))
			assistantParts = append(assistantParts, chat.Part{
				Kind:     chat.PartToolCall,
				ToolCall: &chat.ToolCall{ID: tc.ID, Name: tc.Name, Input: tc.Input},
			})

			// Send tool_call SSE to client for UI
			emit(sseToolCall{
				Type: "tool_call",
				Name: tc.Name,
				Args: json.RawMessage(tc.Input),
			})
		}

		// Append assistant tool_use message to request
		assistantMsg := chat.Message{
			Role:  chat.RoleAssistant,
			Parts: assistantParts,
		}
		req.Messages = append(req.Messages, assistantMsg)

		// Persist assistant tool_use message
		persistParts(db, chatID, "assistant", assistantParts, logger)

		// Execute tools and build tool_result parts
		tctx := toolContext{db: db, paperID: paperID, pdfPath: pdfPath}
		var resultParts []chat.Part
		for _, tc := range toolCalls {
			toolStart := time.Now()
			result := executeToolCall(tc.Name, string(tc.Input), tctx, logger)
			logger.Info("tool_result",
				"name", tc.Name,
				"content_type", result.contentType,
				"result_length", len(result.text),
				"duration", time.Since(toolStart),
			)

			part := chat.Part{
				Kind: chat.PartToolResult,
				ToolResult: &chat.ToolResult{
					ToolCallID: tc.ID,
				},
			}
			sse := sseToolResult{
				Type: "tool_result",
				Name: tc.Name,
			}

			if result.contentType == "image" {
				part.ToolResult.Image = result.image
				sse.ContentType = "image"
				sse.ImageData = result.imageData
				sse.Text = fmt.Sprintf("Rendered page %s as image", string(tc.Input))
				sse.Preview = "Page snapshot rendered"
			} else {
				part.ToolResult.Content = result.text
				sse.Text = result.text
				sse.Preview = truncatePreview(result.text, toolResultPreviewLen)
			}

			resultParts = append(resultParts, part)
			emit(sse)
		}

		// Append user message with tool_result parts
		req.Messages = append(req.Messages, chat.Message{
			Role:  chat.RoleUser,
			Parts: resultParts,
		})

		// Persist user tool_result message
		persistParts(db, chatID, "user", resultParts, logger)
	}

	logger.Info("response_complete",
		"response_length", fullText.Len(),
		"tool_iterations", toolIterations,
		"total_duration", time.Since(responseStart),
	)

	// Store final assistant message
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

// relayEvents reads events from the ActiveStream buffer and writes them as
// SSE to the HTTP response. Returns when the stream completes or the HTTP
// client disconnects.
func relayEvents(httpCtx context.Context, stream *ActiveStream, w http.ResponseWriter, flusher http.Flusher) {
	flush := func() {
		if flusher != nil {
			flusher.Flush()
		}
	}

	offset := 0
	for {
		events, done := stream.EventsSince(offset)
		for _, ev := range events {
			fmt.Fprintf(w, "data: %s\n\n", ev.Data)
			flush()
			offset++
		}
		if done {
			return
		}
		// Capture notify channel, then recheck for events that may have
		// arrived between EventsSince and Notify (avoids lost wakeup).
		ch := stream.Notify()
		if events, done = stream.EventsSince(offset); len(events) > 0 || done {
			continue
		}
		select {
		case <-httpCtx.Done():
			return
		case <-ch:
		}
	}
}

// persistParts saves a message with JSON-serialized chat.Part slices to the DB.
func persistParts(db *sql.DB, chatID, role string, parts []chat.Part, logger *slog.Logger) {
	id, err := newUUID()
	if err != nil {
		logger.Error("failed to generate UUID for tool message", "error", err)
		return
	}
	partsJSON, err := json.Marshal(parts)
	if err != nil {
		logger.Error("failed to marshal content parts", "error", err)
		return
	}
	partsStr := string(partsJSON)
	msg := store.Message{
		ID:            id,
		ChatSessionID: chatID,
		Role:          role,
		ContentBlocks: &partsStr,
		CreatedAt:     time.Now().UTC().Format(time.RFC3339),
	}
	if err := store.CreateMessage(db, msg); err != nil {
		logger.Error("failed to persist tool message", "role", role, "error", err)
	}
}

// buildMultimodalUserMessage creates a domain message with text and image parts from request attachments.
func buildMultimodalUserMessage(content string, attachments []attachment) chat.Message {
	var parts []chat.Part

	var textContent strings.Builder
	textContent.WriteString(content)
	for _, att := range attachments {
		if att.Text != "" {
			textContent.WriteString(fmt.Sprintf("\n\n[Attached region from page %d]\n%s", att.Page, att.Text))
		}
	}
	parts = append(parts, chat.Part{
		Kind: chat.PartText,
		Text: textContent.String(),
	})

	for _, att := range attachments {
		if att.ImageData != "" {
			parts = append(parts, chat.Part{
				Kind:  chat.PartImage,
				Image: &chat.Image{MediaType: "image/png", Data: att.ImageData},
			})
		}
	}

	return chat.Message{Role: chat.RoleUser, Parts: parts}
}

// buildPartsFromPersistedAttachments reads images from disk and builds parts for a past user message.
func buildPartsFromPersistedAttachments(content string, atts []store.Attachment, logger *slog.Logger) ([]chat.Part, error) {
	var parts []chat.Part

	var textContent strings.Builder
	textContent.WriteString(content)
	for _, att := range atts {
		if att.Text != "" {
			textContent.WriteString(fmt.Sprintf("\n\n[Attached region from page %d]\n%s", att.Page, att.Text))
		}
	}
	parts = append(parts, chat.Part{
		Kind: chat.PartText,
		Text: textContent.String(),
	})

	for _, att := range atts {
		imgBytes, err := os.ReadFile(att.FilePath)
		if err != nil {
			logger.Warn("failed to read persisted attachment image", "id", att.ID, "path", att.FilePath, "error", err)
			continue
		}
		parts = append(parts, chat.Part{
			Kind:  chat.PartImage,
			Image: &chat.Image{MediaType: "image/png", Data: base64.StdEncoding.EncodeToString(imgBytes)},
		})
	}

	return parts, nil
}

func handleReconnectStream(registry *StreamRegistry, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		chatID := r.PathValue("chatId")

		stream := registry.Get(chatID)
		if stream == nil {
			writeError(w, http.StatusNotFound, "no active stream", logger)
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		flusher, _ := w.(http.Flusher)
		relayEvents(r.Context(), stream, w, flusher)
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
	Type        string `json:"type"`
	Name        string `json:"name"`
	Text        string `json:"text"`
	Preview     string `json:"preview"`
	ContentType string `json:"content_type,omitempty"`
	ImageData   string `json:"image_data,omitempty"`
}

const toolResultPreviewLen = 200

func truncatePreview(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// toolExecResult holds the result of a tool execution.
type toolExecResult struct {
	text        string
	contentType string     // "text" or "image"
	image       *chat.Image // non-nil for image results
	imageData   string      // base64-encoded image for SSE
}

func textResult(s string) toolExecResult {
	return toolExecResult{text: s, contentType: "text"}
}

func executeToolCall(name, input string, tc toolContext, logger *slog.Logger) toolExecResult {
	switch name {
	case "search_pdf":
		var args struct {
			Query string `json:"query"`
		}
		if err := json.Unmarshal([]byte(input), &args); err != nil {
			return textResult(fmt.Sprintf("Error parsing arguments: %v", err))
		}

		// Try FTS5 index first
		if tc.db != nil && tc.paperID != "" {
			results, err := store.SearchPageText(tc.db, tc.paperID, args.Query)
			if err == nil && len(results) > 0 {
				b, _ := json.Marshal(results)
				return textResult(string(b))
			}
		}

		// Fall back to pdftotext
		results, err := pdf.SearchText(tc.pdfPath, args.Query)
		if err != nil {
			logger.Warn("search_pdf failed", "error", err)
			return textResult(fmt.Sprintf("Error searching PDF: %v", err))
		}
		if len(results) == 0 {
			return textResult("No matches found.")
		}
		b, _ := json.Marshal(results)
		return textResult(string(b))

	case "read_page":
		var args struct {
			Page int `json:"page"`
		}
		if err := json.Unmarshal([]byte(input), &args); err != nil {
			return textResult(fmt.Sprintf("Error parsing arguments: %v", err))
		}

		// Try indexed text first
		if tc.db != nil && tc.paperID != "" {
			text, err := store.GetPageText(tc.db, tc.paperID, args.Page)
			if err == nil {
				return textResult(text)
			}
		}

		// Fall back to pdftotext
		text, err := pdf.ExtractPageText(tc.pdfPath, args.Page)
		if err != nil {
			logger.Warn("read_page failed", "error", err)
			return textResult(fmt.Sprintf("Error reading page: %v", err))
		}
		return textResult(text)

	case "go_to_page":
		var args struct {
			Page int `json:"page"`
		}
		if err := json.Unmarshal([]byte(input), &args); err != nil {
			return textResult(fmt.Sprintf("Error parsing arguments: %v", err))
		}
		return textResult(fmt.Sprintf("Navigated to page %d.", args.Page))

	case "snapshot_page":
		var args struct {
			Page int `json:"page"`
		}
		if err := json.Unmarshal([]byte(input), &args); err != nil {
			return textResult(fmt.Sprintf("Error parsing arguments: %v", err))
		}

		pngBytes, err := pdf.RenderPage(tc.pdfPath, args.Page)
		if err != nil {
			logger.Warn("snapshot_page failed", "error", err, "page", args.Page)
			return textResult(fmt.Sprintf("Error rendering page: %v", err))
		}

		b64 := base64.StdEncoding.EncodeToString(pngBytes)
		return toolExecResult{
			text:        fmt.Sprintf("Rendered page %d as image (%d bytes)", args.Page, len(pngBytes)),
			contentType: "image",
			imageData:   b64,
			image:       &chat.Image{MediaType: "image/png", Data: b64},
		}

	default:
		return textResult(fmt.Sprintf("Unknown tool: %s", name))
	}
}

// toolContext holds context needed for tool execution.
type toolContext struct {
	db      *sql.DB
	paperID string
	pdfPath string
}

func appendViewerContext(content string, currentPage int) string {
	if currentPage == 0 {
		return content
	}

	var b strings.Builder
	b.WriteString(content)
	b.WriteString(fmt.Sprintf("\n\n[Viewer context]\nCurrent page: %d", currentPage))
	return b.String()
}

func mustJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}
