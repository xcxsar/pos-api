package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/xcxsar/pos-api/internal/auth"
)

type userParams struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type userResponse struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func mapUserResponse(id uuid.UUID, created, updated time.Time, email string) userResponse {
	return userResponse{
		ID:        id,
		CreatedAt: created,
		UpdatedAt: updated,
		Email:     email,
	}
}

func (api *API) CreateUser(w http.ResponseWriter, r *http.Request) {
	var params userParams
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid request format")
		return
	}

	dbUser, err := api.UserService.Create(r.Context(), params.Email, params.Password)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	res := mapUserResponse(dbUser.ID, dbUser.CreatedAt, dbUser.UpdatedAt, dbUser.Email)
	respondWithJSON(w, http.StatusCreated, res)
}

func (api *API) GetUserByID(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("userID")
	parsedID, err := uuid.Parse(userID)

	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	dbUser, err := api.UserService.GetByID(r.Context(), parsedID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "no user found")
		return
	}

	res := mapUserResponse(dbUser.ID, dbUser.CreatedAt, dbUser.UpdatedAt, dbUser.Email)
	respondWithJSON(w, http.StatusCreated, res)
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

	dbUser, err := api.UserService.UpdateEmail(r.Context(), userID, param.Email)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	res := mapUserResponse(dbUser.ID, dbUser.CreatedAt, dbUser.UpdatedAt, dbUser.Email)
	respondWithJSON(w, http.StatusCreated, res)
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

	dbUser, err := api.UserService.UpdatePassword(r.Context(), userID, param.Password)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	res := mapUserResponse(dbUser.ID, dbUser.CreatedAt, dbUser.UpdatedAt, dbUser.Email)
	respondWithJSON(w, http.StatusCreated, res)
}
