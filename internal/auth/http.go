package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

type contextKey string

const UserIDKey contextKey = "userID"

func AuthMiddleware(jwtSecret string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token, err := GetBearerToken(r.Header)
		if err != nil {
			http.Error(w, `{"error": "missing or invalid token"}`, http.StatusUnauthorized)
			return
		}

		userID, err := ValidateJWT(token, jwtSecret)
		if err != nil {
			http.Error(w, `{"error": "invalid or expired token"}`, http.StatusUnauthorized)
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
