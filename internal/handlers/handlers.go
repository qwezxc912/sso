package handlers

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/mail"
	"strconv"

	"github.com/gin-gonic/gin"
	errs "github.com/qweq1232/sso/internal/lib/errors"
	"github.com/qweq1232/sso/internal/service"
)

type Handler struct {
	log     *slog.Logger
	service *service.Service
}

type request struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	AppID    int32  `json:"app_id"`
}

func New(log *slog.Logger, service *service.Service) *Handler {
	return &Handler{
		log:     log,
		service: service,
	}
}

func (h *Handler) Register(ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req request

		if err := c.ShouldBindJSON(&req); err != nil {
			h.log.Error("failed to bind request", slog.Any("err", err))

			c.JSON(http.StatusBadRequest, gin.H{

				"error": "invalid request",
			})

			return
		}

		if err := validateRequest(req); err != nil {
			h.log.Error("failed to validate request", slog.Any("err", err))

			c.JSON(http.StatusBadRequest, gin.H{

				"error": "inalid request",
			})

			return
		}

		token, uid, err := h.service.Register(
			ctx,
			req.Password,
			req.Email,
			req.AppID,
		)

		if err != nil {
			if errors.Is(err, errs.AlreadyExists) {
				c.JSON(http.StatusBadRequest, gin.H{
					"status":  "already exists",
					"message": "user already exists",
				})
			}

			h.log.Error("failed to register user", slog.Any("err", err))

			c.JSON(http.StatusInternalServerError, gin.H{

				"error": "internal error",
			})

			return
		}

		c.SetCookie("jwt_token", token, 1500, "/", "localhost", false, true)
		c.SetCookie("uid", strconv.Itoa(uid), 5000, "/", "localhost", false, true)

		c.JSON(http.StatusCreated, gin.H{
			"status": "ok",
		})
	}
}

func (h *Handler) Login(ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req request

		if err := c.ShouldBindJSON(&req); err != nil {
			h.log.Error("failed to bind request", slog.Any("err", err))

			c.JSON(http.StatusBadRequest, gin.H{
				"error": errs.InvalidRequest,
			})

			return
		}

		if err := validateRequest(req); err != nil {
			h.log.Error("failed to validate request", slog.Any("err", err))

			c.JSON(http.StatusBadRequest, gin.H{
				"error": errs.InvalidRequest,
			})

			return
		}

		token, err := h.service.Login(
			ctx,
			req.Password,
			req.Email,
			req.AppID,
		)
		if err != nil {
			if errors.Is(err, errs.NotFound) {
				c.JSON(http.StatusNotFound, gin.H{
					"status":  "not found",
					"message": "not found",
				})

				return
			}

			h.log.Error("failed to login user", slog.Any("err", err))

			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "internal",
			})

			return
		}

		c.SetCookie("jwt_token", token, 1500, "/", "localhost", false, true)

		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	}
}

func validateRequest(req request) error {
	if _, err := mail.ParseAddress(req.Email); err != nil {
		return errs.InvalidRequest
	}

	if len(req.Password) < 8 {
		return errs.InvalidRequest
	}

	if req.AppID != 1 {
		return errs.InvalidRequest
	}

	return nil
}
