package user

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrRequiredCredentials = errors.New("email and password are required")

type CreateDTO struct {
	Email    string
	Password string
}

type UpdateEmailDTO struct {
	ID    uuid.UUID
	Email string
}

type UpdatePasswordDTO struct {
	ID       uuid.UUID
	Password string
}

type Response struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func toResponse(id uuid.UUID, email string, createdAt, updatedAt time.Time) Response {
	return Response{
		ID:        id,
		Email:     email,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}
}
