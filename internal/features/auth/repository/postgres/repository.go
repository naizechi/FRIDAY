package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	coreerrors "github.com/naizechi/friday/internal/core/errors"
	"github.com/naizechi/friday/internal/core/repository/postgres/pool"
	"github.com/naizechi/friday/internal/features/auth/service"
)

type Repository struct {
	pool pool.Pool
}

func NewRepository(p pool.Pool) *Repository {
	return &Repository{pool: p}
}

func (r *Repository) Create(ctx context.Context, email, passwordHash string) (service.User, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OptTimeout())
	defer cancel()

	db := pool.DBFromContext(ctx, r.pool)

	var model userModel
	err := db.QueryRow(ctx, `
		INSERT INTO users (email, password_hash)
		VALUES ($1, $2)
		RETURNING id, email, password_hash, created_at
	`, email, passwordHash).Scan(
		&model.ID,
		&model.Email,
		&model.PasswordHash,
		&model.CreatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return service.User{}, fmt.Errorf("%w: email already registered", coreerrors.ErrConflict)
		}
		return service.User{}, fmt.Errorf("insert user: %w", err)
	}

	return userToDomain(model), nil
}

func (r *Repository) GetByEmail(ctx context.Context, email string) (service.User, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OptTimeout())
	defer cancel()

	db := pool.DBFromContext(ctx, r.pool)

	var model userModel
	err := db.QueryRow(ctx, `
		SELECT id, email, password_hash, created_at
		FROM users
		WHERE lower(email) = lower($1)
	`, email).Scan(
		&model.ID,
		&model.Email,
		&model.PasswordHash,
		&model.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return service.User{}, fmt.Errorf("%w: user", coreerrors.ErrNotFound)
		}
		return service.User{}, fmt.Errorf("select user by email: %w", err)
	}

	return userToDomain(model), nil
}

func (r *Repository) GetByID(ctx context.Context, id int64) (service.User, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OptTimeout())
	defer cancel()

	db := pool.DBFromContext(ctx, r.pool)

	var model userModel
	err := db.QueryRow(ctx, `
		SELECT id, email, password_hash, created_at
		FROM users
		WHERE id = $1
	`, id).Scan(
		&model.ID,
		&model.Email,
		&model.PasswordHash,
		&model.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return service.User{}, fmt.Errorf("%w: user", coreerrors.ErrNotFound)
		}
		return service.User{}, fmt.Errorf("select user by id: %w", err)
	}

	return userToDomain(model), nil
}

func (r *Repository) Insert(ctx context.Context, userID int64, tokenHash string, expiresAt time.Time) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OptTimeout())
	defer cancel()

	db := pool.DBFromContext(ctx, r.pool)

	_, err := db.Exec(ctx, `
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
	`, userID, tokenHash, expiresAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return fmt.Errorf("%w: refresh token", coreerrors.ErrConflict)
		}
		return fmt.Errorf("insert refresh token: %w", err)
	}

	return nil
}

func (r *Repository) GetByHash(ctx context.Context, tokenHash string) (service.RefreshToken, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OptTimeout())
	defer cancel()

	db := pool.DBFromContext(ctx, r.pool)

	var model refreshTokenModel
	err := db.QueryRow(ctx, `
		SELECT id, user_id, token_hash, expires_at, revoked_at, created_at
		FROM refresh_tokens
		WHERE token_hash = $1
	`, tokenHash).Scan(
		&model.ID,
		&model.UserID,
		&model.TokenHash,
		&model.ExpiresAt,
		&model.RevokedAt,
		&model.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return service.RefreshToken{}, fmt.Errorf("%w: refresh token", coreerrors.ErrNotFound)
		}
		return service.RefreshToken{}, fmt.Errorf("select refresh token: %w", err)
	}

	return refreshTokenToDomain(model), nil
}

func (r *Repository) Revoke(ctx context.Context, id int64) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OptTimeout())
	defer cancel()

	db := pool.DBFromContext(ctx, r.pool)

	tag, err := db.Exec(ctx, `
		UPDATE refresh_tokens
		SET revoked_at = now()
		WHERE id = $1 AND revoked_at IS NULL
	`, id)
	if err != nil {
		return fmt.Errorf("revoke refresh token: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("%w: refresh token", coreerrors.ErrNotFound)
	}

	return nil
}
