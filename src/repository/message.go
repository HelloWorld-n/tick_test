package repository

import (
	"fmt"
	"tick_test/types"
	"tick_test/utils/errDefs"
)

func SaveMessage(msg *types.Message) error {
	if database == nil {
		return errDefs.ErrDatabaseOffline
	}

	var fromId, toId int
	err := database.QueryRow(`SELECT id FROM account WHERE username = $1`, msg.From).Scan(&fromId)
	if err != nil {
		return fmt.Errorf("could not resolve sender id: %w", err)
	}
	err = database.QueryRow(`SELECT id FROM account WHERE username = $1`, msg.To).Scan(&toId)
	if err != nil {
		return fmt.Errorf("could not resolve recipient id: %w", err)
	}

	query := `INSERT INTO messages (from_user, to_user, content, created_at) VALUES ($1, $2, $3, $4)`
	_, err = database.Exec(query, fromId, toId, msg.Content, msg.When)
	return err
}

func FindMessages(username string, sent bool, recv bool) (msgs []types.Message, err error) {
	if database == nil {
		return nil, errDefs.ErrDatabaseOffline
	}

	var userId int
	err = database.QueryRow(`SELECT id FROM account WHERE username = $1`, username).Scan(&userId)
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

	rows, err := database.Query(query, userId)
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
		err = database.QueryRow(`SELECT username FROM account WHERE id = $1`, fromId).Scan(&msg.From)
		if err := rows.Scan(&fromId, &toId, &msg.Content, &msg.When); err != nil {
			return nil, err
		}
		err = database.QueryRow(`SELECT username FROM account WHERE id = $1`, toId).Scan(&msg.To)
		if err := rows.Scan(&fromId, &toId, &msg.Content, &msg.When); err != nil {
			return nil, err
		}
		msgs = append(msgs, msg)
	}
	return msgs, nil
}

func doPostgresPreparationForMessages() {
	if database != nil {
		result, err := database.Exec(`
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

		var columnType string
		err = database.QueryRow(`
			SELECT data_type 
			FROM information_schema.columns 
			WHERE table_name = 'messages' AND column_name = 'from_user';
		`).Scan(&columnType)
		if err != nil {
			fmt.Println("Error querying column type:", err)
			return
		}
		if columnType == "character varying" {
			fmt.Println(result, err)
			res, err := database.Exec(`
				ALTER TABLE messages ADD COLUMN IF NOT EXISTS from_id INT;
			`)
			fmt.Println(res, err)

			res, err = database.Exec(`
				ALTER TABLE messages ADD COLUMN IF NOT EXISTS to_id INT;
			`)
			fmt.Println(res, err)

			res, err = database.Exec(`
				UPDATE messages SET from_id = (
					SELECT id FROM account WHERE account.username = messages.from_user
				);
			`)
			fmt.Println(res, err)

			res, err = database.Exec(`
				UPDATE messages SET to_id = (
					SELECT id FROM account WHERE account.username = messages.to_user
				);
			`)
			fmt.Println(res, err)

			res, err = database.Exec(`ALTER TABLE messages DROP CONSTRAINT IF EXISTS messages_from_user_fkey;`)
			fmt.Println(res, err)
			res, err = database.Exec(`ALTER TABLE messages DROP CONSTRAINT IF EXISTS messages_to_user_fkey;`)
			fmt.Println(res, err)

			res, err = database.Exec(`ALTER TABLE messages DROP COLUMN IF EXISTS from_user;`)
			fmt.Println(res, err)
			res, err = database.Exec(`ALTER TABLE messages DROP COLUMN IF EXISTS to_user;`)
			fmt.Println(res, err)

			res, err = database.Exec(`ALTER TABLE messages RENAME COLUMN from_id TO from_user;`)
			fmt.Println(res, err)
			res, err = database.Exec(`ALTER TABLE messages RENAME COLUMN to_id TO to_user;`)
			fmt.Println(res, err)

			res, err = database.Exec(`
				ALTER TABLE messages 
				ADD CONSTRAINT messages_from_user_fk FOREIGN KEY (from_user) REFERENCES account(id),
				ADD CONSTRAINT messages_to_user_fk FOREIGN KEY (to_user) REFERENCES account(id);
			`)
			fmt.Println(res, err)
		}
	}
}
