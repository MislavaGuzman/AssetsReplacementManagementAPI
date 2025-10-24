package store

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

var (
	ErrNotFound          = errors.New("resource not found")
	ErrConflict          = errors.New("resource already exists")
	QueryTimeoutDuration = time.Second * 5
)

type Storage struct {
	Tickets interface {
		GetAll(ctx context.Context, stage int) ([]AssetReplacementTicket, error)
		GetByID(ctx context.Context, id int64) (*AssetReplacementTicket, error)
		Create(ctx context.Context, ticket *AssetReplacementTicket) error
		Update(ctx context.Context, ticket *AssetReplacementTicket) error
		Delete(ctx context.Context, id int64) error
	}
}

func NewStorage(db *sql.DB) Storage {
	return Storage{
		Tickets: &TicketStore{db: db},
	}
}

func withTx(db *sql.DB, ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}
