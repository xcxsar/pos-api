package user

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrRequiredCredentials = errors.New("email and password are required")
	ErrEmailAlreadyExists  = errors.New("email already exists")
	ErrCouldNotSave        = errors.New("could not save user to the database")
	ErrCouldNotUpdateEmail = errors.New("could not update user email to the database")
	ErrCouldNotUpdatePass  = errors.New("could not update user password to the database")
)

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
