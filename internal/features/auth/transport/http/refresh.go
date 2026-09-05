package http

import (
	"net/http"

	"github.com/naizechi/friday/internal/core/logger"
	"github.com/naizechi/friday/internal/core/transport/http/request"
	"github.com/naizechi/friday/internal/core/transport/http/response"
)

func (h *Handler) refresh(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())
	rh := response.NewHandler(log, w)

	var req refreshRequest
	if err := request.DecodeAndValidateRequest(r, &req); err != nil {
		rh.Error(err, "invalid refresh request")
		return
	}

	session, err := h.svc.Refresh(r.Context(), req.RefreshToken)
	if err != nil {
		rh.Error(err, "token refresh failed")
		return
	}

	rh.OK(sessionToDTO(session))
}
