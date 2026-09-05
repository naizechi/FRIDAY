package http

import (
	"net/http"

	"github.com/naizechi/friday/internal/core/logger"
	"github.com/naizechi/friday/internal/core/transport/http/response"
)

func (h *Handler) me(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())
	rh := response.NewHandler(log, w)

	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		rh.Error(errUnauthorized("missing user"), "unauthorized")
		return
	}

	user, err := h.svc.Me(r.Context(), userID)
	if err != nil {
		rh.Error(err, "failed to load current user")
		return
	}

	rh.OK(userToDTO(user))
}
