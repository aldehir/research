package store

import "database/sql"

// ChatSession represents a chat session associated with a paper.
type ChatSession struct {
	ID        string `json:"id"`
	PaperID   string `json:"paper_id"`
	Title     string `json:"title"`
	CreatedAt string `json:"created_at"`
}

// Message represents a single message within a chat session.
type Message struct {
	ID            string  `json:"id"`
	ChatSessionID string  `json:"chat_session_id"`
	Role          string  `json:"role"`
	Content       string  `json:"content"`
	ContentBlocks *string `json:"content_blocks,omitempty"`
	CreatedAt     string  `json:"created_at"`
}

// Attachment represents a file attached to a message (e.g. a region selection image).
type Attachment struct {
	ID        string `json:"id"`
	MessageID string `json:"message_id"`
	FilePath  string `json:"file_path"`
	Text      string `json:"text"`
	Page      int    `json:"page"`
	CreatedAt string `json:"created_at"`
}

// ChatSessionWithMessages combines a chat session with its messages.
type ChatSessionWithMessages struct {
	ChatSession
	Messages []Message `json:"messages"`
}

// CreateChatSession inserts a new chat session record.
func CreateChatSession(db *sql.DB, s ChatSession) error {
	_, err := db.Exec(
		`INSERT INTO chat_sessions (id, paper_id, title, created_at) VALUES (?, ?, ?, ?)`,
		s.ID, s.PaperID, s.Title, s.CreatedAt,
	)
	return err
}

// ListChatSessions returns all chat sessions for a paper, ordered by created_at descending.
func ListChatSessions(db *sql.DB, paperID string) ([]ChatSession, error) {
	rows, err := db.Query(
		`SELECT id, paper_id, title, created_at FROM chat_sessions WHERE paper_id = ? ORDER BY created_at DESC`,
		paperID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []ChatSession
	for rows.Next() {
		var s ChatSession
		if err := rows.Scan(&s.ID, &s.PaperID, &s.Title, &s.CreatedAt); err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}
	return sessions, rows.Err()
}

// GetChatSession returns a single chat session by ID. Returns sql.ErrNoRows if not found.
func GetChatSession(db *sql.DB, id string) (ChatSession, error) {
	var s ChatSession
	err := db.QueryRow(
		`SELECT id, paper_id, title, created_at FROM chat_sessions WHERE id = ?`, id,
	).Scan(&s.ID, &s.PaperID, &s.Title, &s.CreatedAt)
	return s, err
}

// GetChatSessionWithMessages returns a chat session with all its messages ordered by created_at ascending.
func GetChatSessionWithMessages(db *sql.DB, id string) (ChatSessionWithMessages, error) {
	session, err := GetChatSession(db, id)
	if err != nil {
		return ChatSessionWithMessages{}, err
	}

	messages, err := ListMessages(db, id)
	if err != nil {
		return ChatSessionWithMessages{}, err
	}
	if messages == nil {
		messages = []Message{}
	}

	return ChatSessionWithMessages{
		ChatSession: session,
		Messages:    messages,
	}, nil
}

// DeleteChatSession deletes a chat session by ID. Returns ErrNotFound if it does not exist.
func DeleteChatSession(db *sql.DB, id string) error {
	result, err := db.Exec(`DELETE FROM chat_sessions WHERE id = ?`, id)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// CreateMessage inserts a new message record.
func CreateMessage(db *sql.DB, m Message) error {
	_, err := db.Exec(
		`INSERT INTO messages (id, chat_session_id, role, content, content_blocks, created_at) VALUES (?, ?, ?, ?, ?, ?)`,
		m.ID, m.ChatSessionID, m.Role, m.Content, m.ContentBlocks, m.CreatedAt,
	)
	return err
}

// ListMessages returns all messages for a chat session, ordered by created_at ascending.
func ListMessages(db *sql.DB, chatSessionID string) ([]Message, error) {
	rows, err := db.Query(
		`SELECT id, chat_session_id, role, content, content_blocks, created_at FROM messages WHERE chat_session_id = ? ORDER BY created_at ASC`,
		chatSessionID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var m Message
		if err := rows.Scan(&m.ID, &m.ChatSessionID, &m.Role, &m.Content, &m.ContentBlocks, &m.CreatedAt); err != nil {
			return nil, err
		}
		messages = append(messages, m)
	}
	return messages, rows.Err()
}

// CreateAttachment inserts a new message attachment record.
func CreateAttachment(db *sql.DB, a Attachment) error {
	_, err := db.Exec(
		`INSERT INTO message_attachments (id, message_id, file_path, text, page, created_at) VALUES (?, ?, ?, ?, ?, ?)`,
		a.ID, a.MessageID, a.FilePath, a.Text, a.Page, a.CreatedAt,
	)
	return err
}

// GetAttachment returns a single attachment by ID. Returns sql.ErrNoRows if not found.
func GetAttachment(db *sql.DB, id string) (Attachment, error) {
	var a Attachment
	err := db.QueryRow(
		`SELECT id, message_id, file_path, text, page, created_at FROM message_attachments WHERE id = ?`, id,
	).Scan(&a.ID, &a.MessageID, &a.FilePath, &a.Text, &a.Page, &a.CreatedAt)
	return a, err
}

// ListAttachmentsByMessage returns all attachments for a message.
func ListAttachmentsByMessage(db *sql.DB, messageID string) ([]Attachment, error) {
	rows, err := db.Query(
		`SELECT id, message_id, file_path, text, page, created_at FROM message_attachments WHERE message_id = ? ORDER BY created_at ASC`,
		messageID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var atts []Attachment
	for rows.Next() {
		var a Attachment
		if err := rows.Scan(&a.ID, &a.MessageID, &a.FilePath, &a.Text, &a.Page, &a.CreatedAt); err != nil {
			return nil, err
		}
		atts = append(atts, a)
	}
	return atts, rows.Err()
}

// ListAttachmentsByChat returns all attachments for all messages in a chat session.
func ListAttachmentsByChat(db *sql.DB, chatSessionID string) ([]Attachment, error) {
	rows, err := db.Query(
		`SELECT a.id, a.message_id, a.file_path, a.text, a.page, a.created_at
		FROM message_attachments a
		JOIN messages m ON a.message_id = m.id
		WHERE m.chat_session_id = ?
		ORDER BY a.created_at ASC`,
		chatSessionID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var atts []Attachment
	for rows.Next() {
		var a Attachment
		if err := rows.Scan(&a.ID, &a.MessageID, &a.FilePath, &a.Text, &a.Page, &a.CreatedAt); err != nil {
			return nil, err
		}
		atts = append(atts, a)
	}
	return atts, rows.Err()
}
