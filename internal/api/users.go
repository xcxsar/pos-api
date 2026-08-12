package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/xcxsar/pos-api/internal/service"
	"github.com/xcxsar/pos-api/internal/store/sqlc"
)

type userParams struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func mapCreateUserRow(u sqlc.CreateUserRow) User {
	return User{
		ID:        u.ID,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
		Email:     u.Email,
	}
}

func mapGetUserByIDRow(u sqlc.GetUserByIDRow) User {
	return User{
		ID:        u.ID,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
		Email:     u.Email,
	}
}

func mapSQLCUser(u sqlc.User) User {
	return User{
		ID:        u.ID,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
		Email:     u.Email,
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

	if params.Email == "" || params.Password == "" {
		respondWithError(w, http.StatusBadRequest, "email and password are required")
		return
	}

	if !service.CheckPassword(params.Password) {
		respondWithError(w, http.StatusBadRequest, "password must be at least 8 characters long, contain at lest one uppercase letter, at leat one lowercase letter, at least one digit and at least one of the following special characters: @$!%*?&")
		return
	}

	hashedPassword, err := service.HashPassword(params.Password)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "could not hash password")
		return
	}

	user, err := api.Cfg.Queries.CreateUser(r.Context(), sqlc.CreateUserParams{
		Email:          params.Email,
		HashedPassword: hashedPassword,
	})

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "could not create user")
		return
	}

	res := mapCreateUserRow(user)

	respondWithJSON(w, http.StatusCreated, res)
}

func (api *API) GetUserByID(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("userID")
	parsedID, err := uuid.Parse(userID)

	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	user, err := api.Cfg.Queries.GetUserByID(r.Context(), parsedID)

	if err != nil {
		respondWithError(w, http.StatusNotFound, "no user found")
		return
	}

	res := mapGetUserByIDRow(user)

	respondWithJSON(w, http.StatusOK, res)
}
