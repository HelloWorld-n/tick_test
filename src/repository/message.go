package repository

import (
	"fmt"
	"tick_test/types"
	"tick_test/utils/errDefs"
)

type messageRepository interface {
	SaveMessage(msg *types.Message)
	FindMessages(username string, sent bool, recv bool) (msgs []types.Message, err error)
}

type MessageHandler struct {
	MessageRepository messageRepository
}

func NewMessageHandler(messageRepo messageRepository) (res *MessageHandler) {
	return &MessageHandler{
		MessageRepository: messageRepo,
	}
}

func (r *Repo) SaveMessage(msg *types.Message) error {
	if r.DB.Conn == nil {
		return errDefs.ErrDatabaseOffline
	}
	query := `INSERT INTO messages (from_user, to_user, content, created_at) VALUES ($1, $2, $3, $4)`
	_, err := r.DB.Conn.Exec(query, msg.From, msg.To, msg.Content, msg.When)
	return err
}

func (r *Repo) FindMessages(username string, sent bool, recv bool) (msgs []types.Message, err error) {
	if r.DB.Conn == nil {
		return nil, errDefs.ErrDatabaseOffline
	}

	var query string
	switch {
	case sent && recv:
		query = `SELECT from_user, to_user, content, created_at FROM messages WHERE to_user = $1 OR from_user = $1`
	case sent:
		query = `SELECT from_user, to_user, content, created_at FROM messages WHERE from_user = $1`
	case recv:
		query = `SELECT from_user, to_user, content, created_at FROM messages WHERE to_user = $1`
	default:
		return []types.Message{}, nil
	}

	rows, err := r.DB.Conn.Query(query, username)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	msgs = make([]types.Message, 0)
	for rows.Next() {
		var msg types.Message
		if err := rows.Scan(&msg.From, &msg.To, &msg.Content, &msg.When); err != nil {
			return nil, err
		}
		msgs = append(msgs, msg)
	}
	return msgs, nil
}

func (r *Repo) doPostgresPreparationForMessages() {
	if r.DB.Conn != nil {
		result, err := r.DB.Conn.Exec(`
			CREATE TABLE IF NOT EXISTS messages (
				id SERIAL PRIMARY KEY,
				from_user VARCHAR(100) NOT NULL,
				to_user VARCHAR(100) NOT NULL,
				content TEXT NOT NULL, 
				created_at varchar(30) NOT NULL,
				FOREIGN KEY (from_user) REFERENCES account(username),
				FOREIGN KEY (to_user) REFERENCES account(username)
			);
		`)
		fmt.Println(result, err)
	}
}
