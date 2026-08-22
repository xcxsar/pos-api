package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/xcxsar/pos-api/internal/category"
	"github.com/xcxsar/pos-api/internal/store/sqlc"
)

type categoryParams struct {
	Name string `json:"name"`
}
type categoryResponse struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func mapCategoryResponse(u sqlc.Category) categoryResponse {
	return categoryResponse{
		ID:        u.ID,
		Name:      u.Name,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

func (api *API) CreateCategory(w http.ResponseWriter, r *http.Request) {
	var params categoryParams

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid request format")
		return
	}

	dbCategory, err := api.CategoryService.Create(r.Context(), params.Name)
	if err != nil {
		if errors.Is(err, category.ErrBlankName) || errors.Is(err, category.ErrInvalidCharacters) {
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}
		respondWithError(w, http.StatusInternalServerError, "could not process category registration")
		return
	}

	res := mapCategoryResponse(dbCategory)
	respondWithJSON(w, http.StatusCreated, res)
}

func (api *API) GetCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := api.CategoryService.List(r.Context())

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "could not retrieve categories")
	}

	var res []categoryResponse

	for _, c := range categories {
		mapped := mapCategoryResponse(c)
		res = append(res, mapped)
	}

	respondWithJSON(w, http.StatusOK, res)
}
