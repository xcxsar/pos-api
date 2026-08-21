package api

import (
	"encoding/json"
	"net/http"
)

type LoginRes struct {
	userResponse
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}

func (api *API) LogIn(w http.ResponseWriter, r *http.Request) {
	var params userParams

	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid request format")
		return
	}

	user, token, refreshToken, err := api.AuthService.Login(r.Context(), params.Email, params.Password)

	if err != nil {
		if err.Error() == "incorrect email or password" {
			respondWithError(w, http.StatusUnauthorized, err.Error())
		} else {
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	res := LoginRes{
		userResponse: mapUserResponse(user.ID, user.CreatedAt, user.UpdatedAt, user.Email),
		Token:        token,
		RefreshToken: refreshToken,
	}

	respondWithJSON(w, http.StatusOK, res)
}
