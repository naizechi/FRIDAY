package http

import (
	"net/http"

	"github.com/naizechi/friday/internal/core/logger"
	"github.com/naizechi/friday/internal/core/transport/http/request"
	"github.com/naizechi/friday/internal/core/transport/http/response"
)

func (h *Handler) logout(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())
	rh := response.NewHandler(log, w)

	var req refreshRequest
	if err := request.DecodeAndValidateRequest(r, &req); err != nil {
		rh.Error(err, "invalid logout request")
		return
	}

	if err := h.svc.Logout(r.Context(), req.RefreshToken); err != nil {
		rh.Error(err, "logout failed")
		return
	}

	rh.NoContent()
}
