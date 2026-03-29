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
	ID            string `json:"id"`
	ChatSessionID string `json:"chat_session_id"`
	Role          string `json:"role"`
	Content       string `json:"content"`
	CreatedAt     string `json:"created_at"`
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
		`INSERT INTO messages (id, chat_session_id, role, content, created_at) VALUES (?, ?, ?, ?, ?)`,
		m.ID, m.ChatSessionID, m.Role, m.Content, m.CreatedAt,
	)
	return err
}

// ListMessages returns all messages for a chat session, ordered by created_at ascending.
func ListMessages(db *sql.DB, chatSessionID string) ([]Message, error) {
	rows, err := db.Query(
		`SELECT id, chat_session_id, role, content, created_at FROM messages WHERE chat_session_id = ? ORDER BY created_at ASC`,
		chatSessionID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var m Message
		if err := rows.Scan(&m.ID, &m.ChatSessionID, &m.Role, &m.Content, &m.CreatedAt); err != nil {
			return nil, err
		}
		messages = append(messages, m)
	}
	return messages, rows.Err()
}
