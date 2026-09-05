package http

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	coreerrors "github.com/naizechi/friday/internal/core/errors"
	"github.com/naizechi/friday/internal/core/logger"
	"github.com/naizechi/friday/internal/core/transport/http/response"
)

type contextKey string

const userIDContextKey contextKey = "auth_user_id"

func WithUserID(ctx context.Context, userID int64) context.Context {
	return context.WithValue(ctx, userIDContextKey, userID)
}

func UserIDFromContext(ctx context.Context) (int64, bool) {
	userID, ok := ctx.Value(userIDContextKey).(int64)
	return userID, ok
}

func (h *Handler) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := logger.FromContext(r.Context())
		rh := response.NewHandler(log, w)

		header := r.Header.Get("Authorization")
		if header == "" {
			rh.Error(errUnauthorized("missing authorization header"), "unauthorized")
			return
		}

		const prefix = "Bearer "
		if !strings.HasPrefix(header, prefix) {
			rh.Error(errUnauthorized("invalid authorization header"), "unauthorized")
			return
		}

		token := strings.TrimSpace(strings.TrimPrefix(header, prefix))
		if token == "" {
			rh.Error(errUnauthorized("empty access token"), "unauthorized")
			return
		}

		userID, err := h.svc.ParseAccessToken(token)
		if err != nil {
			rh.Error(err, "unauthorized")
			return
		}

		next.ServeHTTP(w, r.WithContext(WithUserID(r.Context(), userID)))
	})
}

func errUnauthorized(reason string) error {
	return fmt.Errorf("%w: %s", coreerrors.ErrUnauthorized, reason)
}
