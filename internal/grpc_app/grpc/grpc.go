package grpcserv

import (
	"context"
	"fmt"
	"net"

	"github.com/qweq1232/sso/internal/config"
	"github.com/qweq1232/sso/internal/lib/jwts"
	storage "github.com/qweq1232/sso/internal/storage/postgres"
	ssov1 "github.com/qwezxc912/protos/gen/go/qweq1232.sso.v1"
	"google.golang.org/grpc"
)

const (
	appID = "1"
	port  = ":8080"
)

type Server struct {
	config  *config.Config
	storage *storage.Storage
	ssov1.UnimplementedUpdaterServer
}

func New(conf *config.Config, storage *storage.Storage) *Server {
	return &Server{
		conf,
		storage,
		ssov1.UnimplementedUpdaterServer{},
	}
}

func (s *Server) MustRun() {
	l, err := net.Listen("tcp", port)
	if err != nil {
		panic(fmt.Sprintf("failed to run gRPC server: %w", err))
	}

	serv := grpc.NewServer()
	ssov1.RegisterUpdaterServer(serv, s)

	fmt.Println("server is listening")

	if err := serv.Serve(l); err != nil {
		panic(fmt.Sprintf("failed to run grpc server: %w", err))
	}
}

func (s *Server) Update(
	ctx context.Context,
	req *ssov1.UpdateRequest,
) (*ssov1.UpdateResponse, error) {
	email, err := s.storage.GetByID(ctx, req.GetUserId())
	if err != nil {
		return nil, err
	}

	token, err := jwts.CreateToken(
		email,
		req.GetUserId(),
		appID,
		s.config.TokenTTL,
		s.config.SecretKey,
	)
	if err != nil {
		return nil, err
	}

	return &ssov1.UpdateResponse{
		Token: token,
	}, nil
}
