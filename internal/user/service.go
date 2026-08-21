package user

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/xcxsar/pos-api/internal/auth"
	"github.com/xcxsar/pos-api/internal/store/sqlc"
)

type Service struct {
	queries *sqlc.Queries
}

func NewService(q *sqlc.Queries) *Service {
	return &Service{queries: q}
}

func (s *Service) Create(ctx context.Context, email, password string) (sqlc.CreateUserRow, error) {
	if email == "" || password == "" {
		return sqlc.CreateUserRow{}, errors.New("email and password are required")
	}

	if !auth.CheckPassword(password) {
		return sqlc.CreateUserRow{}, errors.New("password must be at least 8 characters long, contain at least one uppercase letter, at least one lowercase letter, at least one digit and at least one special character: @$!%*?&")
	}

	hashedPassword, err := auth.HashPassword(password)

	if err != nil {
		return sqlc.CreateUserRow{}, errors.New("could not hash password")
	}

	return s.queries.CreateUser(ctx, sqlc.CreateUserParams{
		Email:          email,
		HashedPassword: hashedPassword,
	})
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (sqlc.GetUserByIDRow, error) {
	return s.queries.GetUserByID(ctx, id)
}

func (s *Service) UpdateEmail(ctx context.Context, userID uuid.UUID, email string) (sqlc.UpdateUserEmailRow, error) {
	if email == "" {
		return sqlc.UpdateUserEmailRow{}, errors.New("email is required")
	}

	return s.queries.UpdateUserEmail(ctx, sqlc.UpdateUserEmailParams{
		Email: email,
		ID:    userID,
	})
}

func (s *Service) UpdatePassword(ctx context.Context, userID uuid.UUID, password string) (sqlc.UpdateUserPasswordRow, error) {
	if !auth.CheckPassword(password) {
		return sqlc.UpdateUserPasswordRow{}, errors.New("password must be at least 8 characters long, contain at least one uppercase letter, at least one lowercase letter, at least one digit and at least one special character: @$!%*?&")
	}

	hashedPassword, err := auth.HashPassword(password)
	if err != nil {
		return sqlc.UpdateUserPasswordRow{}, errors.New("could not hash password")
	}

	return s.queries.UpdateUserPassword(ctx, sqlc.UpdateUserPasswordParams{
		HashedPassword: hashedPassword,
		ID:             userID,
	})
}
