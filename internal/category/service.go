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

func (s *Service) Create(ctx context.Context, dto CreateDTO) (sqlc.Category, error) {
	if dto.Name == "" {
		return sqlc.Category{}, ErrBlankName
	}

	if !validateName(dto.Name) {
		return sqlc.Category{}, ErrInvalidCharacters
	}

	return s.queries.CreateCategory(ctx, dto.Name)
}

func (s *Service) List(ctx context.Context) ([]sqlc.Category, error) {
	return s.queries.GetCategories(ctx)
}

func (s *Service) GetByID(ctx context.Context, id int64) (sqlc.Category, error) {
	return s.queries.GetCategoryByID(ctx, id)
}

func (s *Service) Update(ctx context.Context, dto UpdateDTO) (sqlc.Category, error) {
	if dto.Name == "" {
		return sqlc.Category{}, ErrBlankName
	}

	if !validateName(dto.Name) {
		return sqlc.Category{}, ErrInvalidCharacters
	}

	return s.queries.UpdateCategory(ctx, sqlc.UpdateCategoryParams{
		ID:   dto.ID,
		Name: dto.Name,
	})
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	return s.queries.DeleteCategory(ctx, id)
}
