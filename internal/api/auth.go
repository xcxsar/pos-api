package api

import (
	"encoding/json"
	"net/http"

	"github.com/xcxsar/pos-api/internal/user"
)

type LoginRes struct {
	user.Response
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}

func (api *API) LogIn(w http.ResponseWriter, r *http.Request) {
	var params userParams

	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid request format")
		return
	}

	dbUser, token, refreshToken, err := api.AuthService.Login(r.Context(), params.Email, params.Password)

	if err != nil {
		if err.Error() == "incorrect email or password" {
			respondWithError(w, http.StatusUnauthorized, err.Error())
		} else {
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	res := LoginRes{
		Response: user.Response{
			ID:        dbUser.ID,
			Email:     dbUser.Email,
			CreatedAt: dbUser.CreatedAt,
			UpdatedAt: dbUser.UpdatedAt,
		},
		Token:        token,
		RefreshToken: refreshToken,
	}

	respondWithJSON(w, http.StatusOK, res)
}
