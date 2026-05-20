package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/qweq1232/sso/internal/config"
	"github.com/qweq1232/sso/internal/handlers"
	"github.com/qweq1232/sso/internal/http_app/http"
	"github.com/qweq1232/sso/internal/service"
	storage "github.com/qweq1232/sso/internal/storage/postgres"
)

type App struct {
	log      *slog.Logger
	HTTPServ *http.Server
}

func MustNew(
	ctx context.Context,
	log *slog.Logger,
	cfg *config.Config,
) *App {
	db, err := storage.New(ctx, cfg.DSN)
	if err != nil {
		panic(fmt.Sprintf("failed to init storage: %w", err))
	}

	service := service.New(db, db, cfg.TokenTTL, cfg.SecretKey)

	handler := handlers.New(log, service)

	httpServ := http.New(log, db, handler, cfg.Serv.Port)

	return &App{
		log:      log,
		HTTPServ: httpServ,
	}
}
