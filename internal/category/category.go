package category

import (
	"errors"
	"time"
	"unicode"
)

var (
	ErrInvalidCharacters = errors.New("category name cannot contain numbers, spaces or special characters")
)

type CreateDTO struct {
	Name string
}

type UpdateDTO struct {
	ID   int64
	Name string
}

type Response struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func validateName(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, r := range s {
		if !unicode.IsLetter(r) {
			return false
		}
	}
	return true
}
