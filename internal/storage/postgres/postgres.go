package storage

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	emptyValue = 0
)

type Storage struct {
	pool *pgxpool.Pool
}

func New(ctx context.Context, dsn string) (*Storage, error) {
	const op = "internal.storage.postgres.New"

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer tx.Rollback(ctx)

	stmt := `
		CREATE TABLE IF NOT EXISTS users (
		uid SERIAL PRIMARY KEY,
		email VARCHAR NOT NULL UNIQUE,
		passhash BYTEA NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP);
	`

	if _, err = tx.Exec(ctx, stmt); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{pool: pool}, nil
}

func (s *Storage) Shutdown() {
	s.pool.Close()
}

func (s *Storage) Create(
	ctx context.Context,
	email string,
	passhash []byte,
) (int32, error) {
	const op = "internal.storage.postgres.Create"

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return emptyValue, fmt.Errorf("%s: %w", op, err)
	}
	defer tx.Rollback(ctx)

	stmt := `
		INSERT INTO users VALUES(email, passhash)
		VALUES ($1, $2) RETURNING uid;
	`

	var uid int32

	err = tx.QueryRow(ctx, stmt, email, passhash).Scan(&uid)
	if err != nil {
		return emptyValue, fmt.Errorf("%s: %w", op, err)
	}

	if err = tx.Commit(ctx); err != nil {
		return emptyValue, fmt.Errorf("%s: %w", op, err)
	}

	return uid, nil
}

func (s *Storage) User(
	ctx context.Context,
	email string,
) (int32, []byte, error) {
	const op = "internal.storage.postgres.User"

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return emptyValue, nil, fmt.Errorf("%s: %w", op, err)
	}
	defer tx.Rollback(ctx)

	var (
		uid      int32
		passhash []byte
	)

	err = tx.QueryRow(ctx, "", `
		SELECT uid, passhash FROM users
		WHERE email = $1;
		`, email,
	).Scan(&uid, &passhash)
	if err != nil {
		return emptyValue, nil, fmt.Errorf("%s: %w", op, err)
	}

	if err = tx.Commit(ctx); err != nil {
		return emptyValue, nil, fmt.Errorf("%s: %w", op, err)
	}

	return uid, passhash, nil
}
