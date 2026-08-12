package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
	"unicode"

	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

type TokenType string

const (
	TokenTypeAccess TokenType = "pos-api-access"
)

func CheckPassword(password string) bool {
	if len(password) < 8 {
		return false
	}

	var hasLower, hasUpper, hasDigit, hasSpecial bool

	for _, char := range password {
		switch {
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsDigit(char):
			hasDigit = true
		case strings.ContainsRune("@$!%*?&", char):
			hasSpecial = true
		}
	}

	return hasLower && hasUpper && hasDigit && hasSpecial
}

func HashPassword(password string) (string, error) {
	hash, err := argon2id.CreateHash(password, argon2id.DefaultParams)

	if err != nil {
		return "", err
	}

	return hash, nil
}

func MatchPassword(password, hash string) (bool, error) {
	match, err := argon2id.ComparePasswordAndHash(password, hash)

	if err != nil {
		return false, err
	}

	return match, nil
}

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	issuer := string(TokenTypeAccess)
	issuedAt := jwt.NewNumericDate(time.Now().UTC())
	expiresAt := jwt.NewNumericDate(time.Now().UTC().Add(expiresIn))
	subject := userID.String()
	claims := jwt.RegisteredClaims{Issuer: issuer, Subject: subject, ExpiresAt: expiresAt, IssuedAt: issuedAt}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString([]byte(tokenSecret))

	if err != nil {
		log.Println("Error signing token")
		return "", err
	}

	return signedToken, nil
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	tokenAddr, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(tokenSecret), nil
	})

	if err != nil {
		return uuid.Nil, err
	}

	claims, ok := tokenAddr.Claims.(*jwt.RegisteredClaims)
	if !ok || !tokenAddr.Valid {
		return uuid.Nil, errors.New("invalid token claims")
	}

	idStr := claims.Subject
	issuer := claims.Issuer

	if issuer != string(TokenTypeAccess) {
		return uuid.Nil, errors.New("invalid issuer")
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		return uuid.Nil, err
	}

	return id, nil
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

func MakeRefreshToken() string {
	token := make([]byte, 32)
	rand.Read(token)
	encodedToken := hex.EncodeToString(token)
	return encodedToken
}
