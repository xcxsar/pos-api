package product

import (
	"errors"

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
