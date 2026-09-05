package healthhttp

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/naizechi/friday/internal/core/logger"
	"github.com/naizechi/friday/internal/core/transport/http/response"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) Register(r chi.Router) {
	r.Get("/health", h.health)
}

func (h *Handler) health(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())
	response.NewHandler(log, w).OK(map[string]string{
		"status": "ok",
	})
}
