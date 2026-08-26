package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/xcxsar/pos-api/internal/category"
)

type categoryParams struct {
	Name string `json:"name"`
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

	res, err := api.CategoryService.Create(r.Context(), dto)
	if err != nil {
		if errors.Is(err, category.ErrBlankName) || errors.Is(err, category.ErrInvalidCharacters) {
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}
		respondWithError(w, http.StatusInternalServerError, "could not process category registration")
		return
	}

	respondWithJSON(w, http.StatusCreated, res)
}

func (api *API) GetCategories(w http.ResponseWriter, r *http.Request) {
	res, err := api.CategoryService.List(r.Context())

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "could not retrieve categories")
		return
	}

	respondWithJSON(w, http.StatusOK, res)
}

func (api *API) GetCategoryByID(w http.ResponseWriter, r *http.Request) {
	CategoryID := r.PathValue("categoryID")
	parsedID, err := strconv.ParseInt(CategoryID, 10, 64)

	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid category ID")
		return
	}

	res, err := api.CategoryService.GetByID(r.Context(), parsedID)

	if err != nil {
		respondWithError(w, http.StatusNotFound, "requested category does not exist")
		return
	}

	respondWithJSON(w, http.StatusOK, res)
}

func (api *API) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	CategoryID := r.PathValue("categoryID")
	parsedID, err := strconv.ParseInt(CategoryID, 10, 64)

	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid category ID")
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

	res, err := api.CategoryService.Update(r.Context(), dto)
	if err != nil {
		if errors.Is(err, category.ErrBlankName) || errors.Is(err, category.ErrInvalidCharacters) {
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}
		respondWithError(w, http.StatusInternalServerError, "failed updating infrastructure records")
		return
	}

	respondWithJSON(w, http.StatusOK, res)
}

func (api *API) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	parsedID, err := strconv.ParseInt(r.PathValue("categoryID"), 10, 64)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid target identifier formatting")
		return
	}

	if err := api.CategoryService.Delete(r.Context(), parsedID); err != nil {
		respondWithError(w, http.StatusInternalServerError, "unable to wipe category entry database records")
		return
	}

	respondWithJSON(w, http.StatusNoContent, struct{}{})
}
