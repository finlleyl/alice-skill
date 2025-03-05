package store

import (
	"context"
	"errors"
	"time"
)

var ErrConflict = errors.New("data conflict")

type Store interface {
	FindRecepient(ctx context.Context, username string) (userID string, err error)
	ListMessages(ctx context.Context, userID string) ([]Message, error)
	GetMessage(ctx context.Context, id int64) (*Message, error)
	SaveMessages(ctx context.Context, userID string, msg Message) error
	RegisterUser(ctx context.Context, userID, username string) error
}

type Message struct {
	ID      int64
	Sender  string
	Time    time.Time
	Payload string
}
