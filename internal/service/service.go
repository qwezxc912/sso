package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	errs "github.com/qweq1232/sso/internal/lib/errors"
	"github.com/qweq1232/sso/internal/lib/jwts"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	up        UserProvider
	uc        UserCreater
	ttlToken  time.Duration
	secretKey string
}

type UserProvider interface {
	User(
		ctx context.Context,
		email string,
	) (int32, []byte, error)
}

type UserCreater interface {
	Create(ctx context.Context,
		email string,
		passhash []byte,
	) (int32, error)
}

func New(
	up UserProvider,
	uc UserCreater,
	ttl time.Duration,
	secretKey string,
) *Service {
	return &Service{
		up:        up,
		uc:        uc,
		ttlToken:  ttl,
		secretKey: secretKey,
	}
}

func (s *Service) Login(
	ctx context.Context,
	pass string,
	email string,
	appID int32,
) (string, error) {
	const op = "service.service.Login"

	uid, passhash, err := s.up.User(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", errs.NotFound
		}

		return "", fmt.Errorf("%s: %w", op, err)
	}

	if err = bcrypt.CompareHashAndPassword(passhash, []byte(pass)); err != nil {
		return "", errs.InvalidRequest
	}

	token, err := jwts.CreateToken(
		email,
		uid,
		strconv.Itoa(int(appID)),
		s.ttlToken,
		s.secretKey,
	)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return token, nil
}

func (s *Service) Register(
	ctx context.Context,
	pass string,
	email string,
	appID int32,
) (string, int, error) {
	const op = "service.service.Register"

	passhash, err := bcrypt.GenerateFromPassword(
		[]byte(pass),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return "", 0, fmt.Errorf("%s: %w", op, err)
	}

	uid, err := s.uc.Create(ctx, email, passhash)
	if err != nil {
		return "", 0, fmt.Errorf("%s: %w", op, err)
	}

	token, err := jwts.CreateToken(
		email,
		uid,
		strconv.Itoa(int(appID)),
		s.ttlToken,
		s.secretKey,
	)
	if err != nil {
		return "", 0, fmt.Errorf("%s: %w", op, err)
	}

	return token, int(uid), nil
}
