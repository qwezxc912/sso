package http

import (
	"context"
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/qweq1232/sso/internal/handlers"
	storage "github.com/qweq1232/sso/internal/storage/postgres"
)

type Server struct {
	log     *slog.Logger
	db      *storage.Storage
	handler *handlers.Handler
	port    string
}

func New(
	log *slog.Logger,
	db *storage.Storage,
	handler *handlers.Handler,
	port string,
) *Server {
	return &Server{
		log:     log,
		db:      db,
		handler: handler,
		port:    port,
	}
}

func (s *Server) Run(ctx context.Context) {
	r := gin.Default()

	r.GET("/", s.handler.Login(ctx))
	r.POST("/", s.handler.Register(ctx))

	r.Run(s.port)
}

func (s *Server) Stop() {
	s.db.Shutdown()
}
