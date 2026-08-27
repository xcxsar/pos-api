package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/xcxsar/pos-api/internal/auth"
)

func (api *API) LogIn(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid request format")
		return
	}

	res, err := api.AuthService.Login(r.Context(), params.Email, params.Password)

	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			respondWithError(w, http.StatusUnauthorized, err.Error())
		} else {
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	respondWithJSON(w, http.StatusOK, res)
}
