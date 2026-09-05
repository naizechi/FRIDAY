package http

import (
	"net/http"

	"github.com/naizechi/friday/internal/core/logger"
	"github.com/naizechi/friday/internal/core/transport/http/request"
	"github.com/naizechi/friday/internal/core/transport/http/response"
)

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())
	rh := response.NewHandler(log, w)

	var req credentialsRequest
	if err := request.DecodeAndValidateRequest(r, &req); err != nil {
		rh.Error(err, "invalid login request")
		return
	}

	session, err := h.svc.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		rh.Error(err, "login failed")
		return
	}

	rh.OK(sessionToDTO(session))
}
