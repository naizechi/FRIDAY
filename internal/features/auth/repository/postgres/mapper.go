package postgres

import "github.com/naizechi/friday/internal/features/auth/service"

func userToDomain(m userModel) service.User {
	return service.User{
		ID:           m.ID,
		Email:        m.Email,
		PasswordHash: m.PasswordHash,
		CreatedAt:    m.CreatedAt,
	}
}

func refreshTokenToDomain(m refreshTokenModel) service.RefreshToken {
	return service.RefreshToken{
		ID:        m.ID,
		UserID:    m.UserID,
		TokenHash: m.TokenHash,
		ExpiresAt: m.ExpiresAt,
		RevokedAt: m.RevokedAt,
		CreatedAt: m.CreatedAt,
	}
}
