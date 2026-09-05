package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"

	coreerrors "github.com/naizechi/friday/internal/core/errors"
	"github.com/naizechi/friday/internal/core/logger"
	"github.com/naizechi/friday/internal/features/auth/service"
	authhttp "github.com/naizechi/friday/internal/features/auth/transport/http"
	"github.com/naizechi/friday/internal/features/auth/transport/http/mocks"
)

func testLogger() *logger.Logger {
	return &logger.Logger{
		Logger: slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelError})),
	}
}

func withLogger(r *http.Request) *http.Request {
	return r.WithContext(logger.WithContext(r.Context(), testLogger()))
}

func sampleUser() service.User {
	return service.User{
		ID:        1,
		Email:     "user@example.com",
		CreatedAt: time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC),
	}
}

func sampleSession() service.Session {
	return service.Session{
		User:         sampleUser(),
		AccessToken:  "access-token",
		RefreshToken: "refresh-token",
		TokenType:    "bearer",
		ExpiresIn:    900,
	}
}

func publicRouter(svc authhttp.AuthService) chi.Router {
	h := authhttp.NewHandler(svc)
	r := chi.NewRouter()
	h.RegisterPublic(r)
	return r
}

func protectedRouter(svc authhttp.AuthService) chi.Router {
	h := authhttp.NewHandler(svc)
	r := chi.NewRouter()
	r.Group(func(r chi.Router) {
		r.Use(h.RequireAuth)
		h.RegisterProtected(r)
	})
	return r
}

func TestRegister_Created(t *testing.T) {
	t.Parallel()

	svc := &mocks.AuthServiceMock{
		RegisterFunc: func(_ context.Context, email, password string) (service.Session, error) {
			require.Equal(t, "user@example.com", email)
			require.Equal(t, "password123", password)
			return sampleSession(), nil
		},
	}

	body, err := json.Marshal(map[string]string{
		"email":    "user@example.com",
		"password": "password123",
	})
	require.NoError(t, err)

	req := withLogger(httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body)))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	publicRouter(svc).ServeHTTP(rr, req)

	require.Equal(t, http.StatusCreated, rr.Code)
	require.Equal(t, "/me", rr.Header().Get("Location"))

	var resp map[string]any
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	require.Equal(t, "access-token", resp["access_token"])
	require.Equal(t, "refresh-token", resp["refresh_token"])
}

func TestRegister_InvalidBody(t *testing.T) {
	t.Parallel()

	svc := &mocks.AuthServiceMock{}
	body, err := json.Marshal(map[string]string{
		"email":    "not-an-email",
		"password": "short",
	})
	require.NoError(t, err)

	req := withLogger(httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body)))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	publicRouter(svc).ServeHTTP(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestRegister_Conflict(t *testing.T) {
	t.Parallel()

	svc := &mocks.AuthServiceMock{
		RegisterFunc: func(context.Context, string, string) (service.Session, error) {
			return service.Session{}, coreerrors.ErrConflict
		},
	}

	body, err := json.Marshal(map[string]string{
		"email":    "user@example.com",
		"password": "password123",
	})
	require.NoError(t, err)

	req := withLogger(httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body)))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	publicRouter(svc).ServeHTTP(rr, req)

	require.Equal(t, http.StatusConflict, rr.Code)
}

func TestLogin_OK(t *testing.T) {
	t.Parallel()

	svc := &mocks.AuthServiceMock{
		LoginFunc: func(context.Context, string, string) (service.Session, error) {
			return sampleSession(), nil
		},
	}

	body, err := json.Marshal(map[string]string{
		"email":    "user@example.com",
		"password": "password123",
	})
	require.NoError(t, err)

	req := withLogger(httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body)))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	publicRouter(svc).ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
}

func TestLogin_Unauthorized(t *testing.T) {
	t.Parallel()

	svc := &mocks.AuthServiceMock{
		LoginFunc: func(context.Context, string, string) (service.Session, error) {
			return service.Session{}, coreerrors.ErrUnauthorized
		},
	}

	body, err := json.Marshal(map[string]string{
		"email":    "user@example.com",
		"password": "password123",
	})
	require.NoError(t, err)

	req := withLogger(httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body)))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	publicRouter(svc).ServeHTTP(rr, req)

	require.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestRefresh_OK(t *testing.T) {
	t.Parallel()

	svc := &mocks.AuthServiceMock{
		RefreshFunc: func(_ context.Context, token string) (service.Session, error) {
			require.Equal(t, "refresh-token", token)
			return sampleSession(), nil
		},
	}

	body, err := json.Marshal(map[string]string{"refresh_token": "refresh-token"})
	require.NoError(t, err)

	req := withLogger(httptest.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewReader(body)))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	publicRouter(svc).ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
}

func TestLogout_NoContent(t *testing.T) {
	t.Parallel()

	svc := &mocks.AuthServiceMock{
		LogoutFunc: func(_ context.Context, token string) error {
			require.Equal(t, "refresh-token", token)
			return nil
		},
	}

	body, err := json.Marshal(map[string]string{"refresh_token": "refresh-token"})
	require.NoError(t, err)

	req := withLogger(httptest.NewRequest(http.MethodPost, "/auth/logout", bytes.NewReader(body)))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	publicRouter(svc).ServeHTTP(rr, req)

	require.Equal(t, http.StatusNoContent, rr.Code)
}

func TestRequireAuth_MissingHeader(t *testing.T) {
	t.Parallel()

	svc := &mocks.AuthServiceMock{}
	req := withLogger(httptest.NewRequest(http.MethodGet, "/me", nil))
	rr := httptest.NewRecorder()
	protectedRouter(svc).ServeHTTP(rr, req)

	require.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestMe_OK(t *testing.T) {
	t.Parallel()

	svc := &mocks.AuthServiceMock{
		ParseAccessTokenFunc: func(token string) (int64, error) {
			require.Equal(t, "good-token", token)
			return 1, nil
		},
		MeFunc: func(_ context.Context, userID int64) (service.User, error) {
			require.Equal(t, int64(1), userID)
			return sampleUser(), nil
		},
	}

	req := withLogger(httptest.NewRequest(http.MethodGet, "/me", nil))
	req.Header.Set("Authorization", "Bearer good-token")
	rr := httptest.NewRecorder()
	protectedRouter(svc).ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	var resp map[string]any
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	require.Equal(t, float64(1), resp["id"])
	require.Equal(t, "user@example.com", resp["email"])
}
