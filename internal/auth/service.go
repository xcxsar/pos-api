package auth

import (
	"context"
	"errors"
	"time"

	"github.com/xcxsar/pos-api/internal/password"
	"github.com/xcxsar/pos-api/internal/store/sqlc"
	"github.com/xcxsar/pos-api/internal/user"
)

var (
	ErrRequiredCredentials = errors.New("email and password are required")
	ErrInvalidCredentials  = errors.New("incorrect email or password")
	ErrTokenGeneration     = errors.New("could not generate access token")
	ErrRefreshTokenSave    = errors.New("could not save refresh token")
)

type Service struct {
	queries   *sqlc.Queries
	jwtSecret string
}

func NewService(queries *sqlc.Queries, jwtSecret string) *Service {
	return &Service{queries: queries, jwtSecret: jwtSecret}
}

type LoginResponse struct {
	user.Response
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}

func (s *Service) Login(ctx context.Context, email, pwd string) (LoginResponse, error) {
	if email == "" || pwd == "" {
		return LoginResponse{}, ErrRequiredCredentials
	}

	dbUser, err := s.queries.GetUserByEmail(ctx, email)
	if err != nil {
		return LoginResponse{}, ErrInvalidCredentials
	}

	match, err := password.Match(pwd, dbUser.HashedPassword)
	if err != nil || !match {
		return LoginResponse{}, ErrInvalidCredentials
	}

	const oneHour = time.Duration(3600) * time.Second
	token, err := MakeJWT(dbUser.ID, s.jwtSecret, oneHour)
	if err != nil {
		return LoginResponse{}, ErrTokenGeneration
	}

	refreshToken := MakeRefreshToken()
	_, err = s.queries.CreateRefreshToken(ctx, sqlc.CreateRefreshTokenParams{
		Token:  refreshToken,
		UserID: dbUser.ID,
	})
	if err != nil {
		return LoginResponse{}, ErrRefreshTokenSave
	}

	res := LoginResponse{
		Response: user.Response{
			ID:        dbUser.ID,
			Email:     dbUser.Email,
			CreatedAt: dbUser.CreatedAt,
			UpdatedAt: dbUser.UpdatedAt,
		},
		Token:        token,
		RefreshToken: refreshToken,
	}

	return res, nil
}
