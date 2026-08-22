package category

import (
	"context"

	"github.com/xcxsar/pos-api/internal/store/sqlc"
)

type Service struct {
	queries *sqlc.Queries
}

func NewService(q *sqlc.Queries) *Service {
	return &Service{queries: q}
}

func (s *Service) Create(ctx context.Context, name string) (sqlc.Category, error) {
	if name == "" {
		return sqlc.Category{}, ErrBlankName
	}

	if !validateName(name) {
		return sqlc.Category{}, ErrInvalidCharacters
	}

	return s.queries.CreateCategory(ctx, name)
}

func (s *Service) List(ctx context.Context) ([]sqlc.Category, error) {
	return s.queries.GetCategories(ctx)
}

func (s *Service) GetByID(ctx context.Context, id int64) (sqlc.Category, error) {
	return s.queries.GetCategoryByID(ctx, id)
}

func (s *Service) Update(ctx context.Context, name string) (sqlc.Category, error) {
	if name == "" {
		return sqlc.Category{}, ErrBlankName
	}

	if !validateName(name) {
		return sqlc.Category{}, ErrInvalidCharacters
	}

	return s.queries.UpdateCategory(ctx, name)
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	return s.queries.DeleteCategory(ctx, id)
}
