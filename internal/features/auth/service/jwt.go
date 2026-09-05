package service

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"

	coreerrors "github.com/naizechi/friday/internal/core/errors"
)

type TokenManager struct {
	secret     []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewTokenManager(cfg Config) *TokenManager {
	return &TokenManager{
		secret:     []byte(cfg.JWTSecret),
		accessTTL:  cfg.AccessTTL,
		refreshTTL: cfg.RefreshTTL,
	}
}

func (m *TokenManager) IssueAccess(userID int64) (string, time.Time, error) {
	now := time.Now().UTC()
	expiresAt := now.Add(m.accessTTL)

	claims := jwt.MapClaims{
		"sub": strconv.FormatInt(userID, 10),
		"iat": now.Unix(),
		"exp": expiresAt.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(m.secret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("sign access token: %w", err)
	}

	return signed, expiresAt, nil
}

func (m *TokenManager) ParseAccess(tokenString string) (int64, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("%w: unexpected signing method", coreerrors.ErrUnauthorized)
		}

		return m.secret, nil
	})
	if err != nil || !token.Valid {
		return 0, fmt.Errorf("%w: invalid access token", coreerrors.ErrUnauthorized)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, fmt.Errorf("%w: invalid access token claims", coreerrors.ErrUnauthorized)
	}

	sub, err := claims.GetSubject()
	if err != nil || sub == "" {
		return 0, fmt.Errorf("%w: missing subject", coreerrors.ErrUnauthorized)
	}

	userID, err := strconv.ParseInt(sub, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%w: invalid subject", coreerrors.ErrUnauthorized)
	}

	return userID, nil
}

func (m *TokenManager) NewRefreshToken() (plain string, hash string, expiresAt time.Time, err error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", "", time.Time{}, fmt.Errorf("generate refresh token: %w", err)
	}

	plain = base64.RawURLEncoding.EncodeToString(raw)
	hash = HashRefreshToken(plain)
	expiresAt = time.Now().UTC().Add(m.refreshTTL)

	return plain, hash, expiresAt, nil
}

func HashRefreshToken(plain string) string {
	sum := sha256.Sum256([]byte(plain))
	return hex.EncodeToString(sum[:])
}

func (m *TokenManager) AccessTTLSeconds() int64 {
	return int64(m.accessTTL.Seconds())
}
