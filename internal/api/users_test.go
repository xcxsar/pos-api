package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/xcxsar/pos-api/internal/auth"
	"github.com/xcxsar/pos-api/internal/password"
	"github.com/xcxsar/pos-api/internal/store/sqlc"
	"github.com/xcxsar/pos-api/internal/user"
)

func newTestUsersAPI(t *testing.T) (*API, sqlmock.Sqlmock, *http.ServeMux) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	svc := user.NewService(sqlc.New(db))
	app := NewAPI(nil, nil, svc, nil, nil)

	const secret = "test-secret"
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/users", app.CreateUser)
	mux.HandleFunc("GET /api/users/{userID}", app.GetUserByID)
	mux.HandleFunc("PUT /api/users/email", auth.AuthMiddleware(secret, app.UpdateUserEmail))
	mux.HandleFunc("PUT /api/users/password", auth.AuthMiddleware(secret, app.UpdateUserPassword))

	return app, mock, mux
}

func doAuthedRequest(mux *http.ServeMux, method, path, body, token string) *httptest.ResponseRecorder {
	var bodyReader io.Reader
	if body != "" {
		bodyReader = bytes.NewBufferString(body)
	}
	req := httptest.NewRequest(method, path, bodyReader)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	return rec
}

func validToken(t *testing.T, userID uuid.UUID) string {
	t.Helper()
	token, err := auth.MakeJWT(userID, "test-secret", time.Minute)
	if err != nil {
		t.Fatalf("MakeJWT failed: %v", err)
	}
	return token
}

func assertUserBody(t *testing.T, rec *httptest.ResponseRecorder, wantEmail string) {
	t.Helper()
	var res user.Response
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("response is not valid JSON: %v (body: %s)", err, rec.Body.String())
	}
	if res.Email != wantEmail {
		t.Errorf("email = %q, want %q", res.Email, wantEmail)
	}
}

func TestCreateUser_Success(t *testing.T) {
	_, mock, mux := newTestUsersAPI(t)
	id := uuid.New()
	now := time.Now().UTC().Truncate(time.Microsecond)

	mock.ExpectQuery(`INSERT INTO users`).
		WithArgs("user@example.com", sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "email"}).
			AddRow(id, now, now, "user@example.com"))

	rec := doRequest(mux, http.MethodPost, "/api/users", `{"email":"user@example.com","password":"TestPass1!"}`)

	assertStatus(t, rec, http.StatusCreated)
	assertUserBody(t, rec, "user@example.com")
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestCreateUser_InvalidJSON(t *testing.T) {
	_, _, mux := newTestUsersAPI(t)

	rec := doRequest(mux, http.MethodPost, "/api/users", `{"email":"user@example.com"`)

	assertStatus(t, rec, http.StatusBadRequest)
	assertErrorBody(t, rec, "invalid request format")
}

func TestCreateUser_MissingCredentials(t *testing.T) {
	_, _, mux := newTestUsersAPI(t)

	rec := doRequest(mux, http.MethodPost, "/api/users", `{"email":"","password":""}`)

	assertStatus(t, rec, http.StatusBadRequest)
	assertErrorBody(t, rec, user.ErrRequiredCredentials.Error())
}

func TestCreateUser_InvalidPassword(t *testing.T) {
	_, _, mux := newTestUsersAPI(t)

	rec := doRequest(mux, http.MethodPost, "/api/users", `{"email":"user@example.com","password":"weak"}`)

	assertStatus(t, rec, http.StatusBadRequest)
	assertErrorBody(t, rec, password.ErrInvalidPassword.Error())
}

func TestCreateUser_EmailAlreadyExists(t *testing.T) {
	_, mock, mux := newTestUsersAPI(t)

	mock.ExpectQuery(`INSERT INTO users`).
		WithArgs("taken@example.com", sqlmock.AnyArg()).
		WillReturnError(&pq.Error{Code: "23505"})

	rec := doRequest(mux, http.MethodPost, "/api/users", `{"email":"taken@example.com","password":"TestPass1!"}`)

	assertStatus(t, rec, http.StatusConflict)
	assertErrorBody(t, rec, user.ErrEmailAlreadyExists.Error())
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestCreateUser_DBError(t *testing.T) {
	_, mock, mux := newTestUsersAPI(t)

	mock.ExpectQuery(`INSERT INTO users`).
		WithArgs("user@example.com", sqlmock.AnyArg()).
		WillReturnError(sql.ErrConnDone)

	rec := doRequest(mux, http.MethodPost, "/api/users", `{"email":"user@example.com","password":"TestPass1!"}`)

	assertStatus(t, rec, http.StatusInternalServerError)
	assertErrorBody(t, rec, user.ErrCouldNotSave.Error())
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestGetUserByID_Success(t *testing.T) {
	_, mock, mux := newTestUsersAPI(t)
	id := uuid.New()
	now := time.Now().UTC().Truncate(time.Microsecond)

	mock.ExpectQuery(`SELECT .+ FROM users WHERE id`).
		WithArgs(id).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "email"}).
			AddRow(id, now, now, "user@example.com"))

	rec := doRequest(mux, http.MethodGet, "/api/users/"+id.String(), "")

	assertStatus(t, rec, http.StatusOK)
	assertUserBody(t, rec, "user@example.com")
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestGetUserByID_InvalidID(t *testing.T) {
	_, _, mux := newTestUsersAPI(t)

	rec := doRequest(mux, http.MethodGet, "/api/users/not-a-uuid", "")

	assertStatus(t, rec, http.StatusBadRequest)
	assertErrorBody(t, rec, "invalid user ID")
}

func TestGetUserByID_NotFound(t *testing.T) {
	_, mock, mux := newTestUsersAPI(t)
	id := uuid.New()

	mock.ExpectQuery(`SELECT .+ FROM users WHERE id`).
		WithArgs(id).
		WillReturnError(sql.ErrNoRows)

	rec := doRequest(mux, http.MethodGet, "/api/users/"+id.String(), "")

	assertStatus(t, rec, http.StatusNotFound)
	assertErrorBody(t, rec, "no user found")
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestUpdateUserEmail_Unauthorized(t *testing.T) {
	_, _, mux := newTestUsersAPI(t)

	rec := doAuthedRequest(mux, http.MethodPut, "/api/users/email", `{"email":"new@example.com"}`, "")

	assertStatus(t, rec, http.StatusUnauthorized)
	assertErrorBody(t, rec, "missing or invalid token")
}

func TestUpdateUserEmail_Success(t *testing.T) {
	_, mock, mux := newTestUsersAPI(t)
	id := uuid.New()
	now := time.Now().UTC().Truncate(time.Microsecond)

	mock.ExpectQuery(`UPDATE users SET email`).
		WithArgs("new@example.com", id).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "email"}).
			AddRow(id, now, now, "new@example.com"))

	rec := doAuthedRequest(mux, http.MethodPut, "/api/users/email", `{"email":"new@example.com"}`, validToken(t, id))

	assertStatus(t, rec, http.StatusOK)
	assertUserBody(t, rec, "new@example.com")
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestUpdateUserEmail_InvalidJSON(t *testing.T) {
	_, _, mux := newTestUsersAPI(t)
	id := uuid.New()

	rec := doAuthedRequest(mux, http.MethodPut, "/api/users/email", `{"email":"new@example.com"`, validToken(t, id))

	assertStatus(t, rec, http.StatusBadRequest)
	assertErrorBody(t, rec, "invalid request format")
}

func TestUpdateUserEmail_AlreadyExists(t *testing.T) {
	_, mock, mux := newTestUsersAPI(t)
	id := uuid.New()

	mock.ExpectQuery(`UPDATE users SET email`).
		WithArgs("taken@example.com", id).
		WillReturnError(&pq.Error{Code: "23505"})

	rec := doAuthedRequest(mux, http.MethodPut, "/api/users/email", `{"email":"taken@example.com"}`, validToken(t, id))

	assertStatus(t, rec, http.StatusConflict)
	assertErrorBody(t, rec, user.ErrEmailAlreadyExists.Error())
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestUpdateUserEmail_DBError(t *testing.T) {
	_, mock, mux := newTestUsersAPI(t)
	id := uuid.New()

	mock.ExpectQuery(`UPDATE users SET email`).
		WithArgs("new@example.com", id).
		WillReturnError(sql.ErrConnDone)

	rec := doAuthedRequest(mux, http.MethodPut, "/api/users/email", `{"email":"new@example.com"}`, validToken(t, id))

	assertStatus(t, rec, http.StatusInternalServerError)
	assertErrorBody(t, rec, user.ErrCouldNotUpdateEmail.Error())
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestUpdateUserPassword_Unauthorized(t *testing.T) {
	_, _, mux := newTestUsersAPI(t)

	rec := doAuthedRequest(mux, http.MethodPut, "/api/users/password", `{"password":"NewPass1!"}`, "")

	assertStatus(t, rec, http.StatusUnauthorized)
	assertErrorBody(t, rec, "missing or invalid token")
}

func TestUpdateUserPassword_Success(t *testing.T) {
	_, mock, mux := newTestUsersAPI(t)
	id := uuid.New()
	now := time.Now().UTC().Truncate(time.Microsecond)

	mock.ExpectQuery(`UPDATE users SET hashed_password`).
		WithArgs(sqlmock.AnyArg(), id).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "email"}).
			AddRow(id, now, now, "user@example.com"))

	rec := doAuthedRequest(mux, http.MethodPut, "/api/users/password", `{"password":"NewPass1!"}`, validToken(t, id))

	assertStatus(t, rec, http.StatusOK)
	assertUserBody(t, rec, "user@example.com")
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestUpdateUserPassword_InvalidJSON(t *testing.T) {
	_, _, mux := newTestUsersAPI(t)
	id := uuid.New()

	rec := doAuthedRequest(mux, http.MethodPut, "/api/users/password", `{"password":"NewPass1!"`, validToken(t, id))

	assertStatus(t, rec, http.StatusBadRequest)
	assertErrorBody(t, rec, "invalid request format")
}

func TestUpdateUserPassword_InvalidPassword(t *testing.T) {
	_, _, mux := newTestUsersAPI(t)
	id := uuid.New()

	rec := doAuthedRequest(mux, http.MethodPut, "/api/users/password", `{"password":"weak"}`, validToken(t, id))

	assertStatus(t, rec, http.StatusBadRequest)
	assertErrorBody(t, rec, password.ErrInvalidPassword.Error())
}

func TestUpdateUserPassword_DBError(t *testing.T) {
	_, mock, mux := newTestUsersAPI(t)
	id := uuid.New()

	mock.ExpectQuery(`UPDATE users SET hashed_password`).
		WithArgs(sqlmock.AnyArg(), id).
		WillReturnError(sql.ErrConnDone)

	rec := doAuthedRequest(mux, http.MethodPut, "/api/users/password", `{"password":"NewPass1!"}`, validToken(t, id))

	assertStatus(t, rec, http.StatusInternalServerError)
	assertErrorBody(t, rec, user.ErrCouldNotUpdatePass.Error())
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}