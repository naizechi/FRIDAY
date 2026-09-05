package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/naizechi/friday/internal/core/logger"

	coreerrors "github.com/naizechi/friday/internal/core/errors"
)

func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	log := logger.FromContextOr(ctx, s.log)

	if refreshToken == "" {
		return fmt.Errorf("%w: refresh token is required", coreerrors.ErrInvalidArgument)
	}

	tokenHash := HashRefreshToken(refreshToken)
	stored, err := s.refreshTokens.GetByHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, coreerrors.ErrNotFound) {
			return nil
		}
		return fmt.Errorf("get refresh token: %w", err)
	}

	if stored.RevokedAt != nil {
		return nil
	}

	if err := s.refreshTokens.Revoke(ctx, stored.ID); err != nil {
		log.Debug("logout failed", "err", err)
		return fmt.Errorf("revoke refresh token: %w", err)
	}

	return nil
}
