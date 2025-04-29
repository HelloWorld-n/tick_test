package repository

import (
	"fmt"
	"tick_test/types"
	"tick_test/utils/errDefs"
)

type MessageRepository interface {
	SaveMessage(msg *types.Message) error
	FindMessages(username string, sent bool, recv bool) (msgs []types.Message, err error)
}

func (r *repo) SaveMessage(msg *types.Message) error {
	if r.DB.Conn == nil {
		return errDefs.ErrDatabaseOffline
	}

	var fromId, toId int
	err := r.DB.Conn.QueryRow(`SELECT id FROM account WHERE username = $1`, msg.From).Scan(&fromId)
	if err != nil {
		return fmt.Errorf("could not resolve sender id: %w", err)
	}
	err = r.DB.Conn.QueryRow(`SELECT id FROM account WHERE username = $1`, msg.To).Scan(&toId)
	if err != nil {
		return fmt.Errorf("could not resolve recipient id: %w", err)
	}

	query := `INSERT INTO messages (from_user, to_user, content, created_at) VALUES ($1, $2, $3, $4)`
	_, err = r.DB.Conn.Exec(query, fromId, toId, msg.Content, msg.When)
	return err
}

func (r *repo) FindMessages(username string, sent bool, recv bool) (msgs []types.Message, err error) {
	if r.DB.Conn == nil {
		return nil, errDefs.ErrDatabaseOffline
	}

	var userId int
	err = r.DB.Conn.QueryRow(`SELECT id FROM account WHERE username = $1`, username).Scan(&userId)
	if err != nil {
		return nil, fmt.Errorf("could not resolve user id: %w", err)
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

	rows, err := r.DB.Conn.Query(query, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var fromId, toId int
	msgs = make([]types.Message, 0)
	for rows.Next() {
		var msg types.Message
		if err := rows.Scan(&fromId, &toId, &msg.Content, &msg.When); err != nil {
			return nil, err
		}
		err = r.DB.Conn.QueryRow(`SELECT username FROM account WHERE id = $1`, fromId).Scan(&msg.From)
		if err := rows.Scan(&fromId, &toId, &msg.Content, &msg.When); err != nil {
			return nil, err
		}
		err = r.DB.Conn.QueryRow(`SELECT username FROM account WHERE id = $1`, toId).Scan(&msg.To)
		if err := rows.Scan(&fromId, &toId, &msg.Content, &msg.When); err != nil {
			return nil, err
		}
		msgs = append(msgs, msg)
	}
	return msgs, nil
}

func (r *repo) doPostgresPreparationForMessages() {
	if r.DB.Conn != nil {
		_, err := r.DB.Conn.Exec(`
			CREATE TABLE IF NOT EXISTS messages (
				id SERIAL PRIMARY KEY,
				from_user BIGINT NOT NULL,
				to_user BIGINT NOT NULL,
				content TEXT NOT NULL, 
				created_at varchar(30) NOT NULL,
				FOREIGN KEY (from_user) REFERENCES account(id),
				FOREIGN KEY (to_user) REFERENCES account(id)
			);
		`)
		logPossibleError(err)
	}
}
