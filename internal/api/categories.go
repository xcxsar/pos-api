package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
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

	dto := category.CreateDTO{
		Name: params.Name,
	}

	dbCategory, err := api.CategoryService.Create(r.Context(), dto)
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

func (api *API) GetCategoryByID(w http.ResponseWriter, r *http.Request) {
	CategoryID := r.PathValue("categoryID")
	parsedID, err := strconv.ParseInt(CategoryID, 10, 64)

	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid product ID")
		return
	}

	dbCategory, err := api.CategoryService.GetByID(r.Context(), parsedID)

	if err != nil {
		respondWithError(w, http.StatusNotFound, "requested product does not exist")
		return
	}

	res := mapCategoryResponse(dbCategory)
	respondWithJSON(w, http.StatusOK, res)
}

func (api *API) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	CategoryID := r.PathValue("categoryID")
	parsedID, err := strconv.ParseInt(CategoryID, 10, 64)

	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid product ID")
		return
	}

	_, err = api.CategoryService.GetByID(r.Context(), parsedID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "targeted category does not exist")
		return
	}

	var params categoryParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid request format")
		return
	}

	dto := category.UpdateDTO{
		ID:   parsedID,
		Name: params.Name,
	}

	updated, err := api.CategoryService.Update(r.Context(), dto)
	if err != nil {
		if errors.Is(err, category.ErrBlankName) || errors.Is(err, category.ErrInvalidCharacters) {
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}
		respondWithError(w, http.StatusInternalServerError, "failed updating infrastructure records")
		return
	}

	res := mapCategoryResponse(updated)
	respondWithJSON(w, http.StatusOK, res)
}
