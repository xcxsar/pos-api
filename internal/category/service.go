package category

import (
	"context"
	"strings"

	"github.com/xcxsar/pos-api/internal/store/sqlc"
)

type Service struct {
	queries *sqlc.Queries
}

func NewService(q *sqlc.Queries) *Service {
	return &Service{queries: q}
}

func toResponse(c sqlc.Category) Response {
	return Response{
		ID:        c.ID,
		Name:      c.Name,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}

func (s *Service) Create(ctx context.Context, dto CreateDTO) (Response, error) {
	name := strings.TrimSpace(dto.Name)
	if name == "" {
		name = "Unnamed"
	}

	if !validateName(name) {
		return Response{}, ErrInvalidCharacters
	}

	c, err := s.queries.CreateCategory(ctx, name)
	if err != nil {
		return Response{}, err
	}

	return toResponse(c), nil
}

func (s *Service) List(ctx context.Context) ([]Response, error) {
	categories, err := s.queries.GetCategories(ctx)
	if err != nil {
		return nil, err
	}

	var res []Response
	for _, c := range categories {
		res = append(res, toResponse(c))
	}

	return res, nil
}

func (s *Service) GetByID(ctx context.Context, id int64) (Response, error) {
	c, err := s.queries.GetCategoryByID(ctx, id)
	if err != nil {
		return Response{}, err
	}

	return toResponse(c), nil
}

func (s *Service) Update(ctx context.Context, dto UpdateDTO) (Response, error) {
	name := strings.TrimSpace(dto.Name)
	if name == "" {
		name = "Unnamed"
	}

	if !validateName(name) {
		return Response{}, ErrInvalidCharacters
	}

	c, err := s.queries.UpdateCategory(ctx, sqlc.UpdateCategoryParams{
		ID:   dto.ID,
		Name: name,
	})
	if err != nil {
		return Response{}, err
	}

	return toResponse(c), nil
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	return s.queries.DeleteCategory(ctx, id)
}
