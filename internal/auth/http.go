package auth

import (
	"errors"
	"net/http"
	"strings"
)

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
