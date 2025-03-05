package pg

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/finlleyl/alice-skill/internal/store"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"strings"
	"time"
)

type Store struct {
	conn *sql.DB
}

func NewStore(conn *sql.DB) *Store {
	return &Store{conn: conn}
}

func (s Store) Bootstrap(ctx context.Context) error {
	tx, err := s.conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	tx.ExecContext(ctx, `
        CREATE TABLE users (
            id varchar(128) PRIMARY KEY,
            username varchar(128)
        )
    `)
	tx.ExecContext(ctx, `CREATE UNIQUE INDEX sender_idx ON users (username)`)

	tx.ExecContext(ctx, `
        CREATE TABLE messages (
            id serial PRIMARY KEY,
            sender varchar(128),
            recepient varchar(128),
            payload text,
            sent_at timestamp with time zone,
            read_at timestamp with time zone DEFAULT NULL
        )
    `)
	tx.ExecContext(ctx, `CREATE INDEX recepient_idx ON messages (recepient)`)

	return tx.Commit()
}

func (s Store) FindRecepient(ctx context.Context, username string) (userID string, err error) {
	row := s.conn.QueryRowContext(ctx, `SELECT id FROM users WHERE username = $1`, username)
	err = row.Scan(&userID)
	return
}

func (s Store) ListMessages(ctx context.Context, userID string) ([]store.Message, error) {
	rows, err := s.conn.QueryContext(ctx, `
        SELECT
            m.id,
            u.username AS sender,
            m.sent_at
        FROM messages m
        JOIN users u ON m.sender = u.id
        WHERE
            m.recepient = $1
    `, userID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []store.Message
	for rows.Next() {
		var m store.Message
		if err := rows.Scan(&m.ID, &m.Sender, &m.Time); err != nil {
			return nil, err
		}
		messages = append(messages, m)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return messages, nil
}

func (s Store) GetMessage(ctx context.Context, id int64) (*store.Message, error) {
	row := s.conn.QueryRowContext(ctx, `
        SELECT
            m.id,
            u.username AS sender,
            m.payload,
            m.sent_at
        FROM messages m
        JOIN users u ON m.sender = u.id
        WHERE
            m.id = $1
    `,
		id,
	)

	var msg store.Message
	err := row.Scan(&msg.ID, &msg.Sender, &msg.Payload, &msg.Time)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

func (s Store) SaveMessages(ctx context.Context, messages ...store.Message) error {
	var values []string
	var args []any
	for i, msg := range messages {
		base := i * 4
		params := fmt.Sprintf("($%d, $%d, $%d, $%d)", base+1, base+2, base+3, base+4)
		values = append(values, params)
		args = append(args, msg.Sender, msg.Recepient, msg.Payload, msg.Time)
	}

	query := `
  INSERT INTO messages
  (sender, recepient, payload, sent_at)
  VALUES ` + strings.Join(values, ",") + `;`

	_, err := s.conn.ExecContext(ctx, query, args...)

	return err
}

func (s Store) RegisterUser(ctx context.Context, userID string, username string) error {
	_, err := s.conn.ExecContext(ctx, `
INSERT INTO users
(id, username) VALUES ($1, $2);`, userID, username)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
			err = store.ErrConflict
		}
	}

	return err
}
