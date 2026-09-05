package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/naizechi/friday/internal/core/logger"

	coreerrors "github.com/naizechi/friday/internal/core/errors"
)

func (s *Service) Login(ctx context.Context, email, password string) (Session, error) {
	log := logger.FromContextOr(ctx, s.log)

	email, err := normalizeEmail(email)
	if err != nil {
		return Session{}, err
	}
	if password == "" {
		return Session{}, fmt.Errorf("%w: password is required", coreerrors.ErrInvalidArgument)
	}

	user, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, coreerrors.ErrNotFound) {
			return Session{}, fmt.Errorf("%w: invalid credentials", coreerrors.ErrUnauthorized)
		}
		return Session{}, fmt.Errorf("get user by email: %w", err)
	}

	if err := CheckPassword(user.PasswordHash, password); err != nil {
		log.Debug("login failed", "err", err)
		return Session{}, err
	}

	session, err := s.issueSession(ctx, user)
	if err != nil {
		return Session{}, err
	}

	return session, nil
}
