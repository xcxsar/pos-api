package product

import (
	"errors"
	"time"

	"github.com/shopspring/decimal"
)

var (
	ErrBlankName    = errors.New("product name cannot be empty")
	ErrInvalidPrice = errors.New("product price cannot be negative")
	ErrInvalidStock = errors.New("product stock cannot be negative")
)

type CreateDTO struct {
	Name       string
	Price      decimal.Decimal
	Stock      int32
	CategoryID *int64
}

type UpdateDTO struct {
	ID         int64
	Name       string
	Price      decimal.Decimal
	Stock      int32
	CategoryID *int64
}

type Response struct {
	ID         int64           `json:"id"`
	Name       string          `json:"name"`
	Price      decimal.Decimal `json:"price"`
	Stock      int32           `json:"stock"`
	CategoryID *int64          `json:"category_id"`
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`
}
