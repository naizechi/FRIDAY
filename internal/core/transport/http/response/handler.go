package response

import (
	"encoding/json"
	"errors"
	"net/http"

	coreerrors "github.com/naizechi/friday/internal/core/errors"
	"github.com/naizechi/friday/internal/core/logger"
)

type Handler struct {
	log *logger.Logger
	rw  http.ResponseWriter
}

func NewHandler(log *logger.Logger, rw http.ResponseWriter) *Handler {
	return &Handler{
		log: log,
		rw:  rw,
	}
}

func (h *Handler) JSON(body any, statusCode int) {
	h.rw.Header().Set("Content-Type", "application/json")
	h.rw.WriteHeader(statusCode)

	if err := json.NewEncoder(h.rw).Encode(body); err != nil {
		h.log.Error("write HTTP response", "err", err)
	}
}

func (h *Handler) OK(data any) {
	h.JSON(data, http.StatusOK)
}

// Created writes a 201 response. An empty location skips the header, which is
// the case for resources without a canonical URL of their own.
func (h *Handler) Created(location string, body any) {
	if location != "" {
		h.rw.Header().Set("Location", location)
	}

	h.JSON(body, http.StatusCreated)
}

func (h *Handler) Accepted(location string, body any) {
	if location != "" {
		h.rw.Header().Set("Location", location)
	}

	h.JSON(body, http.StatusAccepted)
}

func (h *Handler) NoContent() {
	h.rw.WriteHeader(http.StatusNoContent)
}

func (h *Handler) Error(err error, msg string) {
	var (
		statusCode int
		logFn      func(string, ...any)
	)

	switch {
	case errors.Is(err, coreerrors.ErrInvalidArgument):
		statusCode = http.StatusBadRequest
		logFn = h.log.Warn

	case errors.Is(err, coreerrors.ErrNotFound):
		statusCode = http.StatusNotFound
		logFn = h.log.Debug

	case errors.Is(err, coreerrors.ErrConflict):
		statusCode = http.StatusConflict
		logFn = h.log.Warn

	case errors.Is(err, coreerrors.ErrUnauthorized):
		statusCode = http.StatusUnauthorized
		logFn = h.log.Debug

	case errors.Is(err, coreerrors.ErrForbidden):
		statusCode = http.StatusForbidden
		logFn = h.log.Warn

	case coreerrors.IsTimeout(err):
		statusCode = http.StatusGatewayTimeout
		logFn = h.log.Warn

	default:
		statusCode = http.StatusInternalServerError
		logFn = h.log.Error
	}

	logFn(msg, "err", err)

	body := map[string]string{
		"message": msg,
	}
	if statusCode < http.StatusInternalServerError {
		body["error"] = err.Error()
	}

	h.JSON(body, statusCode)
}
