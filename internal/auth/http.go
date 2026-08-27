package auth

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

type contextKey string

const UserIDKey contextKey = "userID"

type errorResponse struct {
	Error string `json:"error"`
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	res, err := json.Marshal(errorResponse{Error: msg})
	if err != nil {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal server error"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(res)
}

func AuthMiddleware(jwtSecret string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token, err := GetBearerToken(r.Header)
		if err != nil {
			respondWithError(w, http.StatusUnauthorized, "missing or invalid token")
			return
		}

		userID, err := ValidateJWT(token, jwtSecret)
		if err != nil {
			respondWithError(w, http.StatusUnauthorized, "invalid or expired token")
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, userID)

		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func GetUserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(UserIDKey).(uuid.UUID)
	return id, ok
}

func GetBearerToken(headers http.Header) (string, error) {
	h := headers.Get("Authorization")

	if h == "" {
		return "", errors.New("missing Authorization header")
	}

	after, found := strings.CutPrefix(h, "Bearer ")

	if found {
		return after, nil
	}

	return "", errors.New("invalid Authorization header")
}
