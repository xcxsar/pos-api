package auth

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/xcxsar/pos-api/internal/password"
	"github.com/xcxsar/pos-api/internal/store/sqlc"
)

const testJWTSecret = "test-secret-key"

func newTestService(t *testing.T) (*Service, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	queries := sqlc.New(db)
	svc := NewService(queries, testJWTSecret)
	return svc, mock
}

func hashedPassword(t *testing.T) string {
	t.Helper()
	hash, err := password.Hash("TestPass1!")
	if err != nil {
		t.Fatalf("failed to hash test password: %v", err)
	}
	return hash
}

func TestLogin_Success(t *testing.T) {
	svc, mock := newTestService(t)
	hash := hashedPassword(t)

	userID := uuid.New()
	now := time.Now().UTC().Truncate(time.Microsecond)

	rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "email", "hashed_password"}).
		AddRow(userID, now, now, "user@example.com", hash)
	mock.ExpectQuery(`SELECT .+ FROM users WHERE email`).WithArgs("user@example.com").WillReturnRows(rows)

	refreshRows := sqlmock.NewRows([]string{"token", "created_at", "updated_at", "user_id", "expires_at", "revoked_at"}).
		AddRow("abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890", now, now, userID, now.Add(60*24*time.Hour), nil)
	mock.ExpectQuery(`INSERT INTO refresh_tokens`).WillReturnRows(refreshRows)

	res, err := svc.Login(context.Background(), "user@example.com", "TestPass1!")
	if err != nil {
		t.Fatalf("Login() unexpected error: %v", err)
	}
	if res.Email != "user@example.com" {
		t.Errorf("res.Email = %q, want %q", res.Email, "user@example.com")
	}
	if !strings.Contains(res.Token, ".") {
		t.Error("expected a valid JWT token")
	}
	if len(res.RefreshToken) != 64 {
		t.Errorf("refreshToken length = %d, want 64", len(res.RefreshToken))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestLogin_EmptyEmail(t *testing.T) {
	svc, _ := newTestService(t)

	_, err := svc.Login(context.Background(), "", "TestPass1!")
	if err == nil {
		t.Fatal("expected error for empty email, got nil")
	}
	if err.Error() != "email and password are required" {
		t.Errorf("error = %q, want %q", err.Error(), "email and password are required")
	}
}

func TestLogin_EmptyPassword(t *testing.T) {
	svc, _ := newTestService(t)

	_, err := svc.Login(context.Background(), "user@example.com", "")
	if err == nil {
		t.Fatal("expected error for empty password, got nil")
	}
	if err.Error() != "email and password are required" {
		t.Errorf("error = %q, want %q", err.Error(), "email and password are required")
	}
}

func TestLogin_BothEmpty(t *testing.T) {
	svc, _ := newTestService(t)

	_, err := svc.Login(context.Background(), "", "")
	if err == nil {
		t.Fatal("expected error for both empty, got nil")
	}
	if err.Error() != "email and password are required" {
		t.Errorf("error = %q, want %q", err.Error(), "email and password are required")
	}
}

func TestLogin_UserNotFound(t *testing.T) {
	svc, mock := newTestService(t)

	mock.ExpectQuery(`SELECT .+ FROM users WHERE email`).
		WithArgs("nobody@example.com").
		WillReturnError(sql.ErrNoRows)

	_, err := svc.Login(context.Background(), "nobody@example.com", "TestPass1!")
	if err == nil {
		t.Fatal("expected error for user not found, got nil")
	}
	if err.Error() != "incorrect email or password" {
		t.Errorf("error = %q, want %q", err.Error(), "incorrect email or password")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	svc, mock := newTestService(t)
	hash := hashedPassword(t)

	userID := uuid.New()
	now := time.Now().UTC().Truncate(time.Microsecond)

	rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "email", "hashed_password"}).
		AddRow(userID, now, now, "user@example.com", hash)
	mock.ExpectQuery(`SELECT .+ FROM users WHERE email`).WithArgs("user@example.com").WillReturnRows(rows)

	_, err := svc.Login(context.Background(), "user@example.com", "WrongPass9!")
	if err == nil {
		t.Fatal("expected error for wrong password, got nil")
	}
	if err.Error() != "incorrect email or password" {
		t.Errorf("error = %q, want %q", err.Error(), "incorrect email or password")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestLogin_CreateRefreshTokenFails(t *testing.T) {
	svc, mock := newTestService(t)
	hash := hashedPassword(t)

	userID := uuid.New()
	now := time.Now().UTC().Truncate(time.Microsecond)

	rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "email", "hashed_password"}).
		AddRow(userID, now, now, "user@example.com", hash)
	mock.ExpectQuery(`SELECT .+ FROM users WHERE email`).WithArgs("user@example.com").WillReturnRows(rows)

	mock.ExpectQuery(`INSERT INTO refresh_tokens`).WillReturnError(sql.ErrConnDone)

	_, err := svc.Login(context.Background(), "user@example.com", "TestPass1!")
	if err == nil {
		t.Fatal("expected error when CreateRefreshToken fails, got nil")
	}
	if err.Error() != "could not save refresh token" {
		t.Errorf("error = %q, want %q", err.Error(), "could not save refresh token")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}
