package service

import (
	"context"
	"time"

	"github.com/naizechi/friday/internal/core/logger"
)

//go:generate go tool moq -pkg mocks -out mocks/users_mock.go . Users
//go:generate go tool moq -pkg mocks -out mocks/refresh_tokens_mock.go . RefreshTokens
//go:generate go tool moq -pkg mocks -out mocks/tx_runner_mock.go . TxRunner
type Users interface {
	Create(ctx context.Context, email, passwordHash string) (User, error)
	GetByEmail(ctx context.Context, email string) (User, error)
	GetByID(ctx context.Context, id int64) (User, error)
}

type RefreshTokens interface {
	Insert(ctx context.Context, userID int64, tokenHash string, expiresAt time.Time) error
	GetByHash(ctx context.Context, tokenHash string) (RefreshToken, error)
	Revoke(ctx context.Context, id int64) error
}

type TxRunner interface {
	WithinTx(ctx context.Context, fn func(ctx context.Context) error) error
}

type User struct {
	ID           int64
	Email        string
	PasswordHash string
	CreatedAt    time.Time
}

type RefreshToken struct {
	ID        int64
	UserID    int64
	TokenHash string
	ExpiresAt time.Time
	RevokedAt *time.Time
	CreatedAt time.Time
}

type Session struct {
	User         User
	AccessToken  string
	RefreshToken string
	TokenType    string
	ExpiresIn    int64
}

type Service struct {
	users         Users
	refreshTokens RefreshTokens
	tx            TxRunner
	tokens        *TokenManager
	log           *logger.Logger
	accessTTL     time.Duration
	refreshTTL    time.Duration
}

func New(
	users Users,
	refreshTokens RefreshTokens,
	tx TxRunner,
	tokens *TokenManager,
	log *logger.Logger,
	cfg Config,
) *Service {
	return &Service{
		users:         users,
		refreshTokens: refreshTokens,
		tx:            tx,
		tokens:        tokens,
		log:           log,
		accessTTL:     cfg.AccessTTL,
		refreshTTL:    cfg.RefreshTTL,
	}
}
