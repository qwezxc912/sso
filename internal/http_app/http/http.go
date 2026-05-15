package http

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/qweq1232/sso/internal/service"
	storage "github.com/qweq1232/sso/internal/storage/postgres"
)

type Server struct {
	log     *slog.Logger
	db      *storage.Storage
	service *service.Service
	port    string
}

func New(
	log *slog.Logger,
	db *storage.Storage,
	service *service.Service,
	port string,
) *Server {
	return &Server{
		log:     log,
		db:      db,
		service: service,
		port:    port,
	}
}

func (s *Server) Run() {
	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "hello",
		})
	})

	r.Run(s.port)
}

func (s *Server) Stop() {
	s.db.Shutdown()
}
