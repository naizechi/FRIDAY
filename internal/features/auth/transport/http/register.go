package http

import (
	"net/http"
	"path"
	"strings"

	"github.com/naizechi/friday/internal/core/logger"
	"github.com/naizechi/friday/internal/core/transport/http/request"
	"github.com/naizechi/friday/internal/core/transport/http/response"
)

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())
	rh := response.NewHandler(log, w)

	var req credentialsRequest
	if err := request.DecodeAndValidateRequest(r, &req); err != nil {
		rh.Error(err, "invalid registration request")
		return
	}

	session, err := h.svc.Register(r.Context(), req.Email, req.Password)
	if err != nil {
		rh.Error(err, "registration failed")
		return
	}

	base := strings.TrimSuffix(r.URL.Path, "/auth/register")
	if base == "" {
		base = "/"
	}
	location := path.Join(base, "me")
	rh.Created(location, sessionToDTO(session))
}
