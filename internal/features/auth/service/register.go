package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/naizechi/friday/internal/core/logger"

	coreerrors "github.com/naizechi/friday/internal/core/errors"
)

func (s *Service) Register(ctx context.Context, email, password string) (Session, error) {
	log := logger.FromContextOr(ctx, s.log)

	email, err := normalizeEmail(email)
	if err != nil {
		return Session{}, err
	}
	if err := validatePassword(password); err != nil {
		return Session{}, err
	}

	passwordHash, err := HashPassword(password)
	if err != nil {
		return Session{}, err
	}

	var session Session
	err = s.tx.WithinTx(ctx, func(txCtx context.Context) error {
		user, createErr := s.users.Create(txCtx, email, passwordHash)
		if createErr != nil {
			return fmt.Errorf("create user: %w", createErr)
		}

		built, issueErr := s.issueSession(txCtx, user)
		if issueErr != nil {
			return issueErr
		}

		session = built
		return nil
	})
	if err != nil {
		log.Debug("register failed", "err", err)
		return Session{}, err
	}

	return session, nil
}

func (s *Service) issueSession(ctx context.Context, user User) (Session, error) {
	accessToken, _, err := s.tokens.IssueAccess(user.ID)
	if err != nil {
		return Session{}, err
	}

	refreshPlain, refreshHash, refreshExpires, err := s.tokens.NewRefreshToken()
	if err != nil {
		return Session{}, err
	}

	if err := s.refreshTokens.Insert(ctx, user.ID, refreshHash, refreshExpires); err != nil {
		return Session{}, fmt.Errorf("insert refresh token: %w", err)
	}

	return Session{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshPlain,
		TokenType:    "bearer",
		ExpiresIn:    s.tokens.AccessTTLSeconds(),
	}, nil
}

func normalizeEmail(email string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(email))
	if normalized == "" || !strings.Contains(normalized, "@") {
		return "", fmt.Errorf("%w: invalid email", coreerrors.ErrInvalidArgument)
	}

	return normalized, nil
}

func validatePassword(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("%w: password must be at least 8 characters", coreerrors.ErrInvalidArgument)
	}

	return nil
}
