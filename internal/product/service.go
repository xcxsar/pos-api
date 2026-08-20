package product

import (
	"context"
	"database/sql"

	"github.com/xcxsar/pos-api/internal/store/sqlc"
)

type Service struct {
	queries *sqlc.Queries
}

func NewService(q *sqlc.Queries) *Service {
	return &Service{queries: q}
}

func (s *Service) Create(ctx context.Context, dto CreateDTO) (sqlc.Product, error) {
	if dto.Name == "" || dto.Name == "Unnamed" {
		return sqlc.Product{}, ErrBlankName
	}
	if dto.Price.IsNegative() {
		return sqlc.Product{}, ErrInvalidPrice
	}
	if dto.Stock < 0 {
		return sqlc.Product{}, ErrInvalidStock
	}

	var categoryID sql.NullInt64
	if dto.CategoryID != nil {
		categoryID = sql.NullInt64{Int64: *dto.CategoryID, Valid: true}
	}

	return s.queries.CreateProduct(ctx, sqlc.CreateProductParams{
		Name:       dto.Name,
		Price:      dto.Price.String(),
		Stock:      dto.Stock,
		CategoryID: categoryID,
	})
}

func (s *Service) List(ctx context.Context) ([]sqlc.Product, error) {
	return s.queries.GetProducts(ctx)
}

func (s *Service) GetByID(ctx context.Context, id int64) (sqlc.Product, error) {
	return s.queries.GetProductById(ctx, id)
}

func (s *Service) Update(ctx context.Context, dto UpdateDTO) (sqlc.Product, error) {
	if dto.Name == "" {
		return sqlc.Product{}, ErrBlankName
	}
	if dto.Price.IsNegative() {
		return sqlc.Product{}, ErrInvalidPrice
	}
	if dto.Stock < 0 {
		return sqlc.Product{}, ErrInvalidStock
	}

	var categoryID sql.NullInt64
	if dto.CategoryID != nil {
		categoryID = sql.NullInt64{Int64: *dto.CategoryID, Valid: true}
	}

	return s.queries.UpdateProduct(ctx, sqlc.UpdateProductParams{
		ID:         dto.ID,
		Name:       dto.Name,
		Price:      dto.Price.String(),
		Stock:      dto.Stock,
		CategoryID: categoryID,
	})
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	return s.queries.DeleteProduct(ctx, id)
}
