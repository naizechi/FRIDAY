package service

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"

	coreerrors "github.com/naizechi/friday/internal/core/errors"
)

const bcryptCost = 12

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}

	return string(hash), nil
}

func CheckPassword(hash, password string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return fmt.Errorf("%w: invalid credentials", coreerrors.ErrUnauthorized)
	}

	return nil
}
