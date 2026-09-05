package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/naizechi/friday/internal/core/logger"

	coreerrors "github.com/naizechi/friday/internal/core/errors"
)

func (s *Service) Refresh(ctx context.Context, refreshToken string) (Session, error) {
	log := logger.FromContextOr(ctx, s.log)

	if refreshToken == "" {
		return Session{}, fmt.Errorf("%w: refresh token is required", coreerrors.ErrInvalidArgument)
	}

	tokenHash := HashRefreshToken(refreshToken)

	var session Session
	err := s.tx.WithinTx(ctx, func(txCtx context.Context) error {
		stored, err := s.refreshTokens.GetByHash(txCtx, tokenHash)
		if err != nil {
			if errors.Is(err, coreerrors.ErrNotFound) {
				return fmt.Errorf("%w: invalid refresh token", coreerrors.ErrUnauthorized)
			}
			return fmt.Errorf("get refresh token: %w", err)
		}

		if stored.RevokedAt != nil {
			return fmt.Errorf("%w: refresh token revoked", coreerrors.ErrUnauthorized)
		}
		if time.Now().UTC().After(stored.ExpiresAt) {
			return fmt.Errorf("%w: refresh token expired", coreerrors.ErrUnauthorized)
		}

		if err := s.refreshTokens.Revoke(txCtx, stored.ID); err != nil {
			return fmt.Errorf("revoke refresh token: %w", err)
		}

		user, err := s.users.GetByID(txCtx, stored.UserID)
		if err != nil {
			if errors.Is(err, coreerrors.ErrNotFound) {
				return fmt.Errorf("%w: invalid refresh token", coreerrors.ErrUnauthorized)
			}
			return fmt.Errorf("get user: %w", err)
		}

		built, err := s.issueSession(txCtx, user)
		if err != nil {
			return err
		}

		session = built
		return nil
	})
	if err != nil {
		log.Debug("refresh failed", "err", err)
		return Session{}, err
	}

	return session, nil
}
