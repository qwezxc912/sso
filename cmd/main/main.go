package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/qweq1232/sso/internal/config"
	grpcapp "github.com/qweq1232/sso/internal/grpc_app"
	app "github.com/qweq1232/sso/internal/http_app"
)

const (
	local = "local"
	prod  = "prod"
	dev   = "dev"
)

func main() {
	cfg := config.MustParseConfig()

	log := setupLogger(cfg.Env)

	ctx := context.Background()

	httpApp := app.MustNew(ctx, log, cfg)
	grpcApp := grpcapp.MustNew(ctx, cfg)

	log.Info("starting application", slog.String("port", cfg.Serv.Port))

	go grpcApp.GRPCServer.MustRun()
	go httpApp.HTTPServ.Run(ctx)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	<-stop

	httpApp.HTTPServ.Stop()

	log.Info("application stopped")
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case local:
		log = slog.New(slog.NewTextHandler(os.Stdout,
			&slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case prod:
		log = slog.New(slog.NewJSONHandler(os.Stdout,
			&slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	case dev:
		log = slog.New(slog.NewJSONHandler(os.Stdout,
			&slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	}

	return log
}
