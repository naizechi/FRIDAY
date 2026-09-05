package postgres

import "time"

type userModel struct {
	ID           int64
	Email        string
	PasswordHash string
	CreatedAt    time.Time
}

type refreshTokenModel struct {
	ID        int64
	UserID    int64
	TokenHash string
	ExpiresAt time.Time
	RevokedAt *time.Time
	CreatedAt time.Time
}
