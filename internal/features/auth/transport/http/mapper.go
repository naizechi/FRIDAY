package http

import "github.com/naizechi/friday/internal/features/auth/service"

func userToDTO(u service.User) userResponse {
	return userResponse{
		ID:        u.ID,
		Email:     u.Email,
		CreatedAt: u.CreatedAt,
	}
}

func sessionToDTO(s service.Session) sessionResponse {
	return sessionResponse{
		User:         userToDTO(s.User),
		AccessToken:  s.AccessToken,
		RefreshToken: s.RefreshToken,
		TokenType:    s.TokenType,
		ExpiresIn:    s.ExpiresIn,
	}
}
