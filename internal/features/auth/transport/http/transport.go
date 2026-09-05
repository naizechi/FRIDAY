package http

import (
	"context"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/naizechi/friday/internal/features/auth/service"
)

//go:generate go tool moq -pkg mocks -out mocks/auth_service_mock.go . AuthService
type AuthService interface {
	Register(ctx context.Context, email, password string) (service.Session, error)
	Login(ctx context.Context, email, password string) (service.Session, error)
	Refresh(ctx context.Context, refreshToken string) (service.Session, error)
	Logout(ctx context.Context, refreshToken string) error
	Me(ctx context.Context, userID int64) (service.User, error)
	ParseAccessToken(token string) (int64, error)
}

type Handler struct {
	svc AuthService
}

func NewHandler(svc AuthService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterPublic(r chi.Router) {
	r.Post("/auth/register", h.register)
	r.Post("/auth/login", h.login)
	r.Post("/auth/refresh", h.refresh)
	r.Post("/auth/logout", h.logout)
}

func (h *Handler) RegisterProtected(r chi.Router) {
	r.Get("/me", h.me)
}

type credentialsRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type userResponse struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

type sessionResponse struct {
	User         userResponse `json:"user"`
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	TokenType    string       `json:"token_type"`
	ExpiresIn    int64        `json:"expires_in"`
}
