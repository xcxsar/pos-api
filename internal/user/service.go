package user

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/xcxsar/pos-api/internal/password"
	"github.com/xcxsar/pos-api/internal/store/sqlc"
)

type Service struct {
	queries *sqlc.Queries
}

func NewService(q *sqlc.Queries) *Service {
	return &Service{queries: q}
}

func (s *Service) Create(ctx context.Context, dto CreateDTO) (Response, error) {
	if dto.Email == "" || dto.Password == "" {
		return Response{}, ErrRequiredCredentials
	}

	if !password.Check(dto.Password) {
		return Response{}, password.ErrInvalidPassword
	}

	hashedPassword, err := password.Hash(dto.Password)

	if err != nil {
		return Response{}, errors.New("could not hash password")
	}

	row, err := s.queries.CreateUser(ctx, sqlc.CreateUserParams{
		Email:          dto.Email,
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

func (s *Service) UpdateEmail(ctx context.Context, dto UpdateEmailDTO) (Response, error) {
	if dto.Email == "" {
		return Response{}, errors.New("email is required")
	}

	row, err := s.queries.UpdateUserEmail(ctx, sqlc.UpdateUserEmailParams{
		Email: dto.Email,
		ID:    dto.ID,
	})
	if err != nil {
		return Response{}, err
	}

	return toResponse(row.ID, row.Email, row.CreatedAt, row.UpdatedAt), nil
}

func (s *Service) UpdatePassword(ctx context.Context, dto UpdatePasswordDTO) (Response, error) {
	if !password.Check(dto.Password) {
		return Response{}, password.ErrInvalidPassword
	}

	hashedPassword, err := password.Hash(dto.Password)
	if err != nil {
		return Response{}, errors.New("could not hash password")
	}

	row, err := s.queries.UpdateUserPassword(ctx, sqlc.UpdateUserPasswordParams{
		HashedPassword: hashedPassword,
		ID:             dto.ID,
	})
	if err != nil {
		return Response{}, err
	}

	return toResponse(row.ID, row.Email, row.CreatedAt, row.UpdatedAt), nil
}
