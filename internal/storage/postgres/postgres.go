package storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	errs "github.com/qweq1232/sso/internal/lib/errors"
)

const (
	emptyValue = 0
)

type Storage struct {
	pool *pgxpool.Pool
}

func New(ctx context.Context, dsn string) (*Storage, error) {
	const op = "internal.storage.postgres.New"

	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	config.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol

	pool, err := pgxpool.NewWithConfig(ctx, config)
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
		INSERT INTO users (email, passhash)
		VALUES ($1, $2) RETURNING uid;
	`

	var uid int32

	err = tx.QueryRow(ctx, stmt, email, passhash).Scan(&uid)
	if err != nil {
		return emptyValue, errs.AlreadyExists
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

	row := tx.QueryRow(
		ctx,
		`SELECT uid, passhash FROM users
		WHERE email = $1;`,
		email,
	)
	if err = row.Scan(&uid, &passhash); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return emptyValue, nil, errs.NotFound
		}

		return emptyValue, nil, fmt.Errorf("%s: %w (%s)", op, err, email)
	}

	if err = tx.Commit(ctx); err != nil {
		return emptyValue, nil, fmt.Errorf("%s: %w", op, err)
	}

	return uid, passhash, nil
}

func (s *Storage) GetByID(
	ctx context.Context,
	id int32,
) (string, error) {
	const op = "internal.storage.postgres.GetByID"

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}
	defer tx.Rollback(ctx)

	var email string

	row := tx.QueryRow(
		ctx,
		`SELECT email FROM users
		WHERE uid = $1;`,
		id,
	)
	if err = row.Scan(&email); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", errs.NotFound
		}

		return "", fmt.Errorf("%s: %w", op, err)
	}

	if err = tx.Commit(ctx); err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return email, nil
}
