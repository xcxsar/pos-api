package api

import (
	"database/sql"
	"net/http"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/xcxsar/pos-api/internal/auth"
	"github.com/xcxsar/pos-api/internal/password"
	"github.com/xcxsar/pos-api/internal/store/sqlc"
)

func newTestAuthAPI(t *testing.T) (*API, sqlmock.Sqlmock, *http.ServeMux) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	svc := auth.NewService(sqlc.New(db), "test-secret")
	app := NewAPI(nil, nil, nil, svc, nil)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/login", app.LogIn)

	return app, mock, mux
}

func loginHash(t *testing.T) string {
	t.Helper()
	hash, err := password.Hash("TestPass1!")
	if err != nil {
		t.Fatalf("failed to hash test password: %v", err)
	}
	return hash
}

func TestLogIn_Success(t *testing.T) {
	_, mock, mux := newTestAuthAPI(t)
	hash := loginHash(t)
	id := uuid.New()
	now := time.Now().UTC().Truncate(time.Microsecond)

	mock.ExpectQuery(`SELECT .+ FROM users WHERE email`).
		WithArgs("user@example.com").
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "email", "hashed_password"}).
			AddRow(id, now, now, "user@example.com", hash))

	mock.ExpectQuery(`INSERT INTO refresh_tokens`).
		WillReturnRows(sqlmock.NewRows([]string{"token", "created_at", "updated_at", "user_id", "expires_at", "revoked_at"}).
			AddRow("abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890", now, now, id, now.Add(60*24*time.Hour), nil))

	rec := doRequest(mux, http.MethodPost, "/api/login", `{"email":"user@example.com","password":"TestPass1!"}`)

	assertStatus(t, rec, http.StatusOK)
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestLogIn_InvalidJSON(t *testing.T) {
	_, _, mux := newTestAuthAPI(t)

	rec := doRequest(mux, http.MethodPost, "/api/login", `{"email":"user@example.com"`)

	assertStatus(t, rec, http.StatusBadRequest)
	assertErrorBody(t, rec, "invalid request format")
}

func TestLogIn_UserNotFound(t *testing.T) {
	_, mock, mux := newTestAuthAPI(t)

	mock.ExpectQuery(`SELECT .+ FROM users WHERE email`).
		WithArgs("nobody@example.com").
		WillReturnError(sql.ErrNoRows)

	rec := doRequest(mux, http.MethodPost, "/api/login", `{"email":"nobody@example.com","password":"TestPass1!"}`)

	assertStatus(t, rec, http.StatusUnauthorized)
	assertErrorBody(t, rec, auth.ErrInvalidCredentials.Error())
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestLogIn_WrongPassword(t *testing.T) {
	_, mock, mux := newTestAuthAPI(t)
	hash := loginHash(t)
	id := uuid.New()
	now := time.Now().UTC().Truncate(time.Microsecond)

	mock.ExpectQuery(`SELECT .+ FROM users WHERE email`).
		WithArgs("user@example.com").
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "email", "hashed_password"}).
			AddRow(id, now, now, "user@example.com", hash))

	rec := doRequest(mux, http.MethodPost, "/api/login", `{"email":"user@example.com","password":"WrongPass9!"}`)

	assertStatus(t, rec, http.StatusUnauthorized)
	assertErrorBody(t, rec, auth.ErrInvalidCredentials.Error())
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestLogIn_EmptyCredentials(t *testing.T) {
	_, _, mux := newTestAuthAPI(t)

	rec := doRequest(mux, http.MethodPost, "/api/login", `{"email":"","password":""}`)

	assertStatus(t, rec, http.StatusInternalServerError)
	assertErrorBody(t, rec, auth.ErrRequiredCredentials.Error())
}