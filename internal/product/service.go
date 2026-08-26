package product

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/shopspring/decimal"
	"github.com/xcxsar/pos-api/internal/store/sqlc"
)

type Service struct {
	queries *sqlc.Queries
}

func NewService(q *sqlc.Queries) *Service {
	return &Service{queries: q}
}

func toResponse(p sqlc.Product) (Response, error) {
	priceDecimal, err := decimal.NewFromString(p.Price)
	if err != nil {
		return Response{}, fmt.Errorf("invalid price format: %w", err)
	}

	var categoryID *int64
	if p.CategoryID.Valid {
		idVal := p.CategoryID.Int64
		categoryID = &idVal
	}

	return Response{
		ID:         p.ID,
		Name:       p.Name,
		Price:      priceDecimal,
		Stock:      p.Stock,
		CategoryID: categoryID,
		CreatedAt:  p.CreatedAt,
		UpdatedAt:  p.UpdatedAt,
	}, nil
}

func (s *Service) Create(ctx context.Context, dto CreateDTO) (Response, error) {
	if dto.Name == "" || dto.Name == "Unnamed" {
		return Response{}, ErrBlankName
	}
	if dto.Price.IsNegative() {
		return Response{}, ErrInvalidPrice
	}
	if dto.Stock < 0 {
		return Response{}, ErrInvalidStock
	}

	var categoryID sql.NullInt64
	if dto.CategoryID != nil {
		categoryID = sql.NullInt64{Int64: *dto.CategoryID, Valid: true}
	}

	p, err := s.queries.CreateProduct(ctx, sqlc.CreateProductParams{
		Name:       dto.Name,
		Price:      dto.Price.String(),
		Stock:      dto.Stock,
		CategoryID: categoryID,
	})
	if err != nil {
		return Response{}, err
	}

	return toResponse(p)
}

func (s *Service) List(ctx context.Context) ([]Response, error) {
	products, err := s.queries.GetProducts(ctx)
	if err != nil {
		return nil, err
	}

	var res []Response
	for _, p := range products {
		r, err := toResponse(p)
		if err != nil {
			return nil, err
		}
		res = append(res, r)
	}

	return res, nil
}

func (s *Service) GetByID(ctx context.Context, id int64) (Response, error) {
	p, err := s.queries.GetProductById(ctx, id)
	if err != nil {
		return Response{}, err
	}

	return toResponse(p)
}

func (s *Service) Update(ctx context.Context, dto UpdateDTO) (Response, error) {
	if dto.Name == "" {
		return Response{}, ErrBlankName
	}
	if dto.Price.IsNegative() {
		return Response{}, ErrInvalidPrice
	}
	if dto.Stock < 0 {
		return Response{}, ErrInvalidStock
	}

	var categoryID sql.NullInt64
	if dto.CategoryID != nil {
		categoryID = sql.NullInt64{Int64: *dto.CategoryID, Valid: true}
	}

	p, err := s.queries.UpdateProduct(ctx, sqlc.UpdateProductParams{
		ID:         dto.ID,
		Name:       dto.Name,
		Price:      dto.Price.String(),
		Stock:      dto.Stock,
		CategoryID: categoryID,
	})
	if err != nil {
		return Response{}, err
	}

	return toResponse(p)
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	return s.queries.DeleteProduct(ctx, id)
}
