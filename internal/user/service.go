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

func (s *Service) Create(ctx context.Context, email, password string) (Response, error) {
	if email == "" || password == "" {
		return Response{}, errors.New("email and password are required")
	}

	if !auth.CheckPassword(password) {
		return Response{}, errors.New("password must be at least 8 characters long, contain at least one uppercase letter, at least one lowercase letter, at least one digit and at least one special character: @$!%*?&")
	}

	hashedPassword, err := auth.HashPassword(password)

	if err != nil {
		return Response{}, errors.New("could not hash password")
	}

	row, err := s.queries.CreateUser(ctx, sqlc.CreateUserParams{
		Email:          email,
		HashedPassword: hashedPassword,
	})
	if err != nil {
		return Response{}, err
	}

	return toResponse(row.ID, row.Email, row.CreatedAt, row.UpdatedAt), nil
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (Response, error) {
	row, err := s.queries.GetUserByID(ctx, id)
	if err != nil {
		return Response{}, err
	}

	return toResponse(row.ID, row.Email, row.CreatedAt, row.UpdatedAt), nil
}

func (s *Service) UpdateEmail(ctx context.Context, userID uuid.UUID, email string) (Response, error) {
	if email == "" {
		return Response{}, errors.New("email is required")
	}

	row, err := s.queries.UpdateUserEmail(ctx, sqlc.UpdateUserEmailParams{
		Email: email,
		ID:    userID,
	})
	if err != nil {
		return Response{}, err
	}

	return toResponse(row.ID, row.Email, row.CreatedAt, row.UpdatedAt), nil
}

func (s *Service) UpdatePassword(ctx context.Context, userID uuid.UUID, password string) (Response, error) {
	if !auth.CheckPassword(password) {
		return Response{}, errors.New("password must be at least 8 characters long, contain at least one uppercase letter, at least one lowercase letter, at least one digit and at least one special character: @$!%*?&")
	}

	hashedPassword, err := auth.HashPassword(password)
	if err != nil {
		return Response{}, errors.New("could not hash password")
	}

	row, err := s.queries.UpdateUserPassword(ctx, sqlc.UpdateUserPasswordParams{
		HashedPassword: hashedPassword,
		ID:             userID,
	})
	if err != nil {
		return Response{}, err
	}

	return toResponse(row.ID, row.Email, row.CreatedAt, row.UpdatedAt), nil
}
