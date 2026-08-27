package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/xcxsar/pos-api/internal/auth"
	"github.com/xcxsar/pos-api/internal/password"
	"github.com/xcxsar/pos-api/internal/user"
)

func (api *API) CreateUser(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid request format")
		return
	}

	res, err := api.UserService.Create(r.Context(), user.CreateDTO{
		Email:    params.Email,
		Password: params.Password,
	})
	if err != nil {
		switch {
		case errors.Is(err, user.ErrRequiredCredentials), errors.Is(err, password.ErrInvalidPassword):
			respondWithError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, user.ErrEmailAlreadyExists):
			respondWithError(w, http.StatusConflict, err.Error())
		default:
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	respondWithJSON(w, http.StatusCreated, res)
}

func (api *API) GetUserByID(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("userID")
	parsedID, err := uuid.Parse(userID)

	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	res, err := api.UserService.GetByID(r.Context(), parsedID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "no user found")
		return
	}

	respondWithJSON(w, http.StatusOK, res)
}

func (api *API) UpdateUserEmail(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var param struct {
		Email string `json:"email"`
	}

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&param)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid request format")
		return
	}

	res, err := api.UserService.UpdateEmail(r.Context(), user.UpdateEmailDTO{
		ID:    userID,
		Email: param.Email,
	})
	if err != nil {
		if errors.Is(err, user.ErrEmailAlreadyExists) {
			respondWithError(w, http.StatusConflict, err.Error())
		} else {
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	respondWithJSON(w, http.StatusOK, res)
}

func (api *API) UpdateUserPassword(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var param struct {
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&param)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid request format")
		return
	}

	res, err := api.UserService.UpdatePassword(r.Context(), user.UpdatePasswordDTO{
		ID:       userID,
		Password: param.Password,
	})
	if err != nil {
		if errors.Is(err, password.ErrInvalidPassword) {
			respondWithError(w, http.StatusBadRequest, err.Error())
		} else {
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	respondWithJSON(w, http.StatusOK, res)
}
