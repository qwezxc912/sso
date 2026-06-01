package grpcapp

import (
	"context"
	"fmt"

	"github.com/qweq1232/sso/internal/config"
	grpcserv "github.com/qweq1232/sso/internal/grpc_app/grpc"
	storage "github.com/qweq1232/sso/internal/storage/postgres"
)

type GRPCApp struct {
	GRPCServer *grpcserv.Server
}

func MustNew(
	ctx context.Context,
	cfg *config.Config,
) *GRPCApp {
	db, err := storage.New(ctx, cfg.DSN)
	if err != nil {
		panic(fmt.Sprintf("failed to init storage: %w", err))
	}

	grpcServ := grpcserv.New(cfg, db)

	return &GRPCApp{
		GRPCServer: grpcServ,
	}
}
