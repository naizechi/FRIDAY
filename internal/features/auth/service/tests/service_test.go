package tests

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	coreerrors "github.com/naizechi/friday/internal/core/errors"
	"github.com/naizechi/friday/internal/core/logger"
	"github.com/naizechi/friday/internal/features/auth/service"
	"github.com/naizechi/friday/internal/features/auth/service/mocks"
)

func testLogger() *logger.Logger {
	return &logger.Logger{
		Logger: slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelError})),
	}
}

func testConfig() service.Config {
	return service.Config{
		JWTSecret:  "test-secret-at-least-32-bytes-long!!",
		AccessTTL:  15 * time.Minute,
		RefreshTTL: 24 * time.Hour,
	}
}

func newTestService(
	users *mocks.UsersMock,
	refresh *mocks.RefreshTokensMock,
	tx *mocks.TxRunnerMock,
) *service.Service {
	cfg := testConfig()
	if tx.WithinTxFunc == nil {
		tx.WithinTxFunc = func(ctx context.Context, fn func(ctx context.Context) error) error {
			return fn(ctx)
		}
	}
	return service.New(users, refresh, tx, service.NewTokenManager(cfg), testLogger(), cfg)
}

func TestRegister_Success(t *testing.T) {
	t.Parallel()

	users := &mocks.UsersMock{
		CreateFunc: func(_ context.Context, email, passwordHash string) (service.User, error) {
			require.Equal(t, "user@example.com", email)
			require.NotEmpty(t, passwordHash)
			return service.User{
				ID:           1,
				Email:        email,
				PasswordHash: passwordHash,
				CreatedAt:    time.Now().UTC(),
			}, nil
		},
	}
	refresh := &mocks.RefreshTokensMock{
		InsertFunc: func(_ context.Context, userID int64, tokenHash string, expiresAt time.Time) error {
			require.Equal(t, int64(1), userID)
			require.NotEmpty(t, tokenHash)
			require.True(t, expiresAt.After(time.Now().UTC()))
			return nil
		},
	}
	tx := &mocks.TxRunnerMock{}

	svc := newTestService(users, refresh, tx)
	session, err := svc.Register(context.Background(), "  User@Example.com ", "password123")

	require.NoError(t, err)
	require.Equal(t, int64(1), session.User.ID)
	require.Equal(t, "user@example.com", session.User.Email)
	require.NotEmpty(t, session.AccessToken)
	require.NotEmpty(t, session.RefreshToken)
	require.Equal(t, "bearer", session.TokenType)
	require.Equal(t, int64(15*60), session.ExpiresIn)
	require.Len(t, users.CreateCalls(), 1)
	require.Len(t, refresh.InsertCalls(), 1)
}

func TestRegister_WeakPassword(t *testing.T) {
	t.Parallel()

	svc := newTestService(&mocks.UsersMock{}, &mocks.RefreshTokensMock{}, &mocks.TxRunnerMock{})
	_, err := svc.Register(context.Background(), "user@example.com", "short")

	require.ErrorIs(t, err, coreerrors.ErrInvalidArgument)
}

func TestRegister_Conflict(t *testing.T) {
	t.Parallel()

	users := &mocks.UsersMock{
		CreateFunc: func(context.Context, string, string) (service.User, error) {
			return service.User{}, coreerrors.ErrConflict
		},
	}
	svc := newTestService(users, &mocks.RefreshTokensMock{}, &mocks.TxRunnerMock{})

	_, err := svc.Register(context.Background(), "user@example.com", "password123")
	require.ErrorIs(t, err, coreerrors.ErrConflict)
}

func TestLogin_Success(t *testing.T) {
	t.Parallel()

	hash, err := service.HashPassword("password123")
	require.NoError(t, err)

	users := &mocks.UsersMock{
		GetByEmailFunc: func(_ context.Context, email string) (service.User, error) {
			require.Equal(t, "user@example.com", email)
			return service.User{
				ID:           7,
				Email:        email,
				PasswordHash: hash,
				CreatedAt:    time.Now().UTC(),
			}, nil
		},
	}
	refresh := &mocks.RefreshTokensMock{
		InsertFunc: func(context.Context, int64, string, time.Time) error { return nil },
	}

	svc := newTestService(users, refresh, &mocks.TxRunnerMock{})
	session, err := svc.Login(context.Background(), "user@example.com", "password123")

	require.NoError(t, err)
	require.Equal(t, int64(7), session.User.ID)
	require.NotEmpty(t, session.AccessToken)
	require.NotEmpty(t, session.RefreshToken)
}

func TestLogin_WrongPassword(t *testing.T) {
	t.Parallel()

	hash, err := service.HashPassword("password123")
	require.NoError(t, err)

	users := &mocks.UsersMock{
		GetByEmailFunc: func(context.Context, string) (service.User, error) {
			return service.User{
				ID:           7,
				Email:        "user@example.com",
				PasswordHash: hash,
			}, nil
		},
	}

	svc := newTestService(users, &mocks.RefreshTokensMock{}, &mocks.TxRunnerMock{})
	_, err = svc.Login(context.Background(), "user@example.com", "wrong-password")
	require.ErrorIs(t, err, coreerrors.ErrUnauthorized)
}

func TestLogin_UnknownEmail(t *testing.T) {
	t.Parallel()

	users := &mocks.UsersMock{
		GetByEmailFunc: func(context.Context, string) (service.User, error) {
			return service.User{}, coreerrors.ErrNotFound
		},
	}

	svc := newTestService(users, &mocks.RefreshTokensMock{}, &mocks.TxRunnerMock{})
	_, err := svc.Login(context.Background(), "missing@example.com", "password123")
	require.ErrorIs(t, err, coreerrors.ErrUnauthorized)
}

func TestRefresh_Rotation(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	oldHash := service.HashRefreshToken("old-refresh-token")

	users := &mocks.UsersMock{
		GetByIDFunc: func(_ context.Context, id int64) (service.User, error) {
			require.Equal(t, int64(3), id)
			return service.User{ID: 3, Email: "user@example.com", CreatedAt: now}, nil
		},
	}
	refresh := &mocks.RefreshTokensMock{
		GetByHashFunc: func(_ context.Context, tokenHash string) (service.RefreshToken, error) {
			require.Equal(t, oldHash, tokenHash)
			return service.RefreshToken{
				ID:        10,
				UserID:    3,
				TokenHash: tokenHash,
				ExpiresAt: now.Add(time.Hour),
			}, nil
		},
		RevokeFunc: func(_ context.Context, id int64) error {
			require.Equal(t, int64(10), id)
			return nil
		},
		InsertFunc: func(_ context.Context, userID int64, tokenHash string, _ time.Time) error {
			require.Equal(t, int64(3), userID)
			require.NotEqual(t, oldHash, tokenHash)
			return nil
		},
	}

	svc := newTestService(users, refresh, &mocks.TxRunnerMock{})
	session, err := svc.Refresh(context.Background(), "old-refresh-token")

	require.NoError(t, err)
	require.Equal(t, int64(3), session.User.ID)
	require.NotEmpty(t, session.AccessToken)
	require.NotEqual(t, "old-refresh-token", session.RefreshToken)
	require.Len(t, refresh.RevokeCalls(), 1)
	require.Len(t, refresh.InsertCalls(), 1)
}

func TestRefresh_Revoked(t *testing.T) {
	t.Parallel()

	revokedAt := time.Now().UTC()
	refresh := &mocks.RefreshTokensMock{
		GetByHashFunc: func(context.Context, string) (service.RefreshToken, error) {
			return service.RefreshToken{
				ID:        10,
				UserID:    3,
				ExpiresAt: time.Now().UTC().Add(time.Hour),
				RevokedAt: &revokedAt,
			}, nil
		},
	}

	svc := newTestService(&mocks.UsersMock{}, refresh, &mocks.TxRunnerMock{})
	_, err := svc.Refresh(context.Background(), "revoked-token")
	require.ErrorIs(t, err, coreerrors.ErrUnauthorized)
}

func TestMe_NotFound(t *testing.T) {
	t.Parallel()

	users := &mocks.UsersMock{
		GetByIDFunc: func(context.Context, int64) (service.User, error) {
			return service.User{}, coreerrors.ErrNotFound
		},
	}

	svc := newTestService(users, &mocks.RefreshTokensMock{}, &mocks.TxRunnerMock{})
	_, err := svc.Me(context.Background(), 99)
	require.ErrorIs(t, err, coreerrors.ErrUnauthorized)
}

func TestLogout_Success(t *testing.T) {
	t.Parallel()

	refresh := &mocks.RefreshTokensMock{
		GetByHashFunc: func(context.Context, string) (service.RefreshToken, error) {
			return service.RefreshToken{
				ID:        5,
				UserID:    1,
				ExpiresAt: time.Now().UTC().Add(time.Hour),
			}, nil
		},
		RevokeFunc: func(_ context.Context, id int64) error {
			require.Equal(t, int64(5), id)
			return nil
		},
	}

	svc := newTestService(&mocks.UsersMock{}, refresh, &mocks.TxRunnerMock{})
	err := svc.Logout(context.Background(), "some-refresh")
	require.NoError(t, err)
	require.Len(t, refresh.RevokeCalls(), 1)
}
