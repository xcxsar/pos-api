package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/xcxsar/pos-api/internal/service"
	"github.com/xcxsar/pos-api/internal/store/sqlc"
)

type LoginRes struct {
	User
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}

func mapLoginRes(u sqlc.User, token, refreshToken string) LoginRes {
	return LoginRes{
		User:         mapSQLCUser(u),
		Token:        token,
		RefreshToken: refreshToken,
	}
}

func (api *API) LogIn(w http.ResponseWriter, r *http.Request) {
	const oneHour = time.Duration(3600) * time.Second

	var params userParams

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&params)

	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	if params.Email == "" || params.Password == "" {
		respondWithError(w, http.StatusBadRequest, "Email or Hashed Password is required")
		return
	}

	user, err := api.Cfg.Queries.GetUserByEmail(r.Context(), params.Email)

	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password")
		return
	}

	match, err := service.MatchPassword(params.Password, user.HashedPassword)

	if err != nil || !match {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password")
		return
	}

	token, err := service.MakeJWT(user.ID, api.Cfg.JWTSecret, oneHour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	refreshToken := service.MakeRefreshToken()
	_, err = api.Cfg.Queries.CreateRefreshToken(r.Context(), sqlc.CreateRefreshTokenParams{
		Token:  refreshToken,
		UserID: user.ID,
	})

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	res := mapLoginRes(user, token, refreshToken)
	res.Token = token

	respondWithJSON(w, http.StatusOK, res)
}
