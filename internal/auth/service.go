package auth

import (
	"context"
	"errors"
	"time"

	"github.com/xcxsar/pos-api/internal/store/sqlc"
)

type Service struct {
	queries   *sqlc.Queries
	jwtSecret string
}

func NewService(queries *sqlc.Queries, jwtSecret string) *Service {
	return &Service{queries: queries, jwtSecret: jwtSecret}
}

func (s *Service) Login(ctx context.Context, email, password string) (sqlc.User, string, string, error) {
	if email == "" || password == "" {
		return sqlc.User{}, "", "", errors.New("email and password are required")
	}

	user, err := s.queries.GetUserByEmail(ctx, email)
	if err != nil {
		return sqlc.User{}, "", "", errors.New("incorrect email or password")
	}

	match, err := MatchPassword(password, user.HashedPassword)
	if err != nil || !match {
		return sqlc.User{}, "", "", errors.New("incorrect email or password")
	}

	const oneHour = time.Duration(3600) * time.Second
	token, err := MakeJWT(user.ID, s.jwtSecret, oneHour)
	if err != nil {
		return sqlc.User{}, "", "", errors.New("could not generate access token")
	}

	refreshToken := MakeRefreshToken()
	_, err = s.queries.CreateRefreshToken(ctx, sqlc.CreateRefreshTokenParams{
		Token:  refreshToken,
		UserID: user.ID,
	})
	if err != nil {
		return sqlc.User{}, "", "", errors.New("could not save refresh token")
	}

	return user, token, refreshToken, nil
}
