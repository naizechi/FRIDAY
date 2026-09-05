package service

import (
	"context"
	"errors"
	"fmt"

	coreerrors "github.com/naizechi/friday/internal/core/errors"
)

func (s *Service) Me(ctx context.Context, userID int64) (User, error) {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, coreerrors.ErrNotFound) {
			return User{}, fmt.Errorf("%w: user not found", coreerrors.ErrUnauthorized)
		}
		return User{}, fmt.Errorf("get user: %w", err)
	}

	return user, nil
}

func (s *Service) ParseAccessToken(token string) (int64, error) {
	return s.tokens.ParseAccess(token)
}
