package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/shopspring/decimal"
	"github.com/xcxsar/pos-api/internal/product"
)

type productParams struct {
	Name       string          `json:"name"`
	Price      decimal.Decimal `json:"price"`
	Stock      int32           `json:"stock"`
	CategoryID *int64          `json:"category_id"`
}

func (api *API) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var params productParams

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid request format")
		return
	}

	dto := product.CreateDTO{
		Name:       params.Name,
		Price:      params.Price,
		Stock:      params.Stock,
		CategoryID: params.CategoryID,
	}

	res, err := api.ProductService.Create(r.Context(), dto)
	if err != nil {
		if errors.Is(err, product.ErrBlankName) || errors.Is(err, product.ErrInvalidPrice) || errors.Is(err, product.ErrInvalidStock) {
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}
		respondWithError(w, http.StatusInternalServerError, "could not process product registration")
		return
	}

	respondWithJSON(w, http.StatusCreated, res)
}

func (api *API) GetProducts(w http.ResponseWriter, r *http.Request) {
	res, err := api.ProductService.List(r.Context())

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "could not retrieve products")
		return
	}

	respondWithJSON(w, http.StatusOK, res)
}

func (api *API) GetProductByID(w http.ResponseWriter, r *http.Request) {
	productID := r.PathValue("productID")
	parsedID, err := strconv.ParseInt(productID, 10, 64)

	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid product ID")
		return
	}

	res, err := api.ProductService.GetByID(r.Context(), parsedID)

	if err != nil {
		respondWithError(w, http.StatusNotFound, "requested product does not exist")
		return
	}

	respondWithJSON(w, http.StatusOK, res)
}

func (api *API) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	productID := r.PathValue("productID")
	parsedID, err := strconv.ParseInt(productID, 10, 64)

	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid product ID")
		return
	}

	_, err = api.ProductService.GetByID(r.Context(), parsedID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "targeted product does not exist")
		return
	}

	var params productParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid parameters format")
		return
	}

	dto := product.UpdateDTO{
		ID:         parsedID,
		Name:       params.Name,
		Price:      params.Price,
		Stock:      params.Stock,
		CategoryID: params.CategoryID,
	}

	res, err := api.ProductService.Update(r.Context(), dto)
	if err != nil {
		if errors.Is(err, product.ErrBlankName) || errors.Is(err, product.ErrInvalidPrice) {
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}
		respondWithError(w, http.StatusInternalServerError, "failed updating infrastructure records")
		return
	}

	respondWithJSON(w, http.StatusOK, res)
}

func (api *API) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	parsedID, err := strconv.ParseInt(r.PathValue("productID"), 10, 64)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid target identifier formatting")
		return
	}

	if err := api.ProductService.Delete(r.Context(), parsedID); err != nil {
		respondWithError(w, http.StatusInternalServerError, "unable to wipe inventory entry database records")
		return
	}

	respondWithJSON(w, http.StatusNoContent, struct{}{})
}
