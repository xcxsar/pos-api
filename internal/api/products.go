package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/shopspring/decimal"
	"github.com/xcxsar/pos-api/internal/store/sqlc"
)

type productParams struct {
	Name       string          `json:"name"`
	Price      decimal.Decimal `json:"price"`
	Stock      int32           `json:"stock"`
	CategoryID *int64          `json:"category_id"`
}

type product struct {
	ID         int64           `json:"id"`
	Name       string          `json:"name"`
	Price      decimal.Decimal `json:"price"`
	Stock      int32           `json:"stock"`
	CategoryID *int64          `json:"category_id"`
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`
}

func mapProduct(u sqlc.Product) (product, error) {
	priceDecimal, err := decimal.NewFromString(u.Price)
	if err != nil {
		return product{}, fmt.Errorf("invalid price format for product %d: %w", u.ID, err)
	}

	var categoryID *int64
	if u.CategoryID.Valid {
		idVal := u.CategoryID.Int64
		categoryID = &idVal
	}

	return product{
		ID:         u.ID,
		Name:       u.Name,
		Price:      priceDecimal,
		Stock:      u.Stock,
		CategoryID: categoryID,
		CreatedAt:  u.CreatedAt,
		UpdatedAt:  u.UpdatedAt,
	}, nil
}

func (api *API) CreateProduct(w http.ResponseWriter, r *http.Request) {
	params := productParams{
		Name:       "Unnamed",
		Price:      decimal.NewFromFloat(0.0),
		Stock:      0,
		CategoryID: nil,
	}

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid request format")
		return
	}

	var categoryID sql.NullInt64
	if params.CategoryID != nil {
		categoryID = sql.NullInt64{
			Int64: *params.CategoryID,
			Valid: true,
		}
	}

	product, err := api.Cfg.Queries.CreateProduct(r.Context(), sqlc.CreateProductParams{
		Name:       params.Name,
		Price:      params.Price.String(),
		Stock:      params.Stock,
		CategoryID: categoryID,
	})

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "could not create product")
		return
	}

	res, err := mapProduct(product)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "could not map product")
		return
	}

	respondWithJSON(w, http.StatusCreated, res)
}
