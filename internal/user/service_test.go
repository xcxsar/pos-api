package user

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/xcxsar/pos-api/internal/password"
	"github.com/xcxsar/pos-api/internal/store/sqlc"
)

var userColumns = []string{"id", "created_at", "updated_at", "email"}

func newTestService(t *testing.T) (*Service, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	queries := sqlc.New(db)
	svc := NewService(queries)
	return svc, mock
}

func userRow(id uuid.UUID, email string, now time.Time) *sqlmock.Rows {
	return sqlmock.NewRows(userColumns).
		AddRow(id, now, now, email)
}

func TestCreate_Success(t *testing.T) {
	svc, mock := newTestService(t)
	id := uuid.New()
	now := time.Now().UTC().Truncate(time.Microsecond)

	mock.ExpectQuery(`INSERT INTO users`).
		WithArgs("user@example.com", sqlmock.AnyArg()).
		WillReturnRows(userRow(id, "user@example.com", now))

	res, err := svc.Create(context.Background(), CreateDTO{
		Email:    "user@example.com",
		Password: "TestPass1!",
	})
	if err != nil {
		t.Fatalf("Create() unexpected error: %v", err)
	}
	if res.ID != id {
		t.Errorf("res.ID = %s, want %s", res.ID, id)
	}
	if res.Email != "user@example.com" {
		t.Errorf("res.Email = %q, want user@example.com", res.Email)
	}
	if res.CreatedAt != now || res.UpdatedAt != now {
		t.Errorf("res timestamps = %+v, want %s", res, now)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestCreate_MissingCredentials(t *testing.T) {
	svc, _ := newTestService(t)

	tests := []struct {
		name string
		dto  CreateDTO
	}{
		{"empty email", CreateDTO{Email: "", Password: "TestPass1!"}},
		{"empty password", CreateDTO{Email: "user@example.com", Password: ""}},
		{"both empty", CreateDTO{Email: "", Password: ""}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.Create(context.Background(), tt.dto)
			if !errors.Is(err, ErrRequiredCredentials) {
				t.Errorf("Create() error = %v, want %v", err, ErrRequiredCredentials)
			}
		})
	}
}

func TestCreate_InvalidPassword(t *testing.T) {
	svc, _ := newTestService(t)

	_, err := svc.Create(context.Background(), CreateDTO{
		Email:    "user@example.com",
		Password: "weak",
	})
	if !errors.Is(err, password.ErrInvalidPassword) {
		t.Errorf("Create() error = %v, want %v", err, password.ErrInvalidPassword)
	}
}

func TestCreate_EmailAlreadyExists(t *testing.T) {
	svc, mock := newTestService(t)

	mock.ExpectQuery(`INSERT INTO users`).
		WithArgs("taken@example.com", sqlmock.AnyArg()).
		WillReturnError(&pq.Error{Code: "23505"})

	_, err := svc.Create(context.Background(), CreateDTO{
		Email:    "taken@example.com",
		Password: "TestPass1!",
	})
	if !errors.Is(err, ErrEmailAlreadyExists) {
		t.Errorf("Create() error = %v, want %v", err, ErrEmailAlreadyExists)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestCreate_DBFails(t *testing.T) {
	svc, mock := newTestService(t)

	mock.ExpectQuery(`INSERT INTO users`).
		WithArgs("user@example.com", sqlmock.AnyArg()).
		WillReturnError(sql.ErrConnDone)

	_, err := svc.Create(context.Background(), CreateDTO{
		Email:    "user@example.com",
		Password: "TestPass1!",
	})
	if !errors.Is(err, ErrCouldNotSave) {
		t.Errorf("Create() error = %v, want %v", err, ErrCouldNotSave)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestGetByID_Success(t *testing.T) {
	svc, mock := newTestService(t)
	id := uuid.New()
	now := time.Now().UTC().Truncate(time.Microsecond)

	mock.ExpectQuery(`SELECT .+ FROM users WHERE id`).
		WithArgs(id).
		WillReturnRows(userRow(id, "user@example.com", now))

	res, err := svc.GetByID(context.Background(), id)
	if err != nil {
		t.Fatalf("GetByID() unexpected error: %v", err)
	}
	if res.ID != id || res.Email != "user@example.com" {
		t.Errorf("res = %+v, want ID %s / user@example.com", res, id)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestGetByID_DBFails(t *testing.T) {
	svc, mock := newTestService(t)
	id := uuid.New()

	mock.ExpectQuery(`SELECT .+ FROM users WHERE id`).
		WithArgs(id).
		WillReturnError(sql.ErrConnDone)

	_, err := svc.GetByID(context.Background(), id)
	if !errors.Is(err, sql.ErrConnDone) {
		t.Errorf("GetByID() error = %v, want sql.ErrConnDone", err)
	}
}

func TestUpdateEmail_Success(t *testing.T) {
	svc, mock := newTestService(t)
	id := uuid.New()
	now := time.Now().UTC().Truncate(time.Microsecond)

	mock.ExpectQuery(`UPDATE users SET email`).
		WithArgs("new@example.com", id).
		WillReturnRows(userRow(id, "new@example.com", now))

	res, err := svc.UpdateEmail(context.Background(), UpdateEmailDTO{
		ID:    id,
		Email: "new@example.com",
	})
	if err != nil {
		t.Fatalf("UpdateEmail() unexpected error: %v", err)
	}
	if res.Email != "new@example.com" {
		t.Errorf("res.Email = %q, want new@example.com", res.Email)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestUpdateEmail_EmptyEmail(t *testing.T) {
	svc, _ := newTestService(t)

	_, err := svc.UpdateEmail(context.Background(), UpdateEmailDTO{
		ID:    uuid.New(),
		Email: "",
	})
	if err == nil {
		t.Fatal("UpdateEmail() expected error for empty email, got nil")
	}
}

func TestUpdateEmail_AlreadyExists(t *testing.T) {
	svc, mock := newTestService(t)
	id := uuid.New()

	mock.ExpectQuery(`UPDATE users SET email`).
		WithArgs("taken@example.com", id).
		WillReturnError(&pq.Error{Code: "23505"})

	_, err := svc.UpdateEmail(context.Background(), UpdateEmailDTO{
		ID:    id,
		Email: "taken@example.com",
	})
	if !errors.Is(err, ErrEmailAlreadyExists) {
		t.Errorf("UpdateEmail() error = %v, want %v", err, ErrEmailAlreadyExists)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestUpdateEmail_DBFails(t *testing.T) {
	svc, mock := newTestService(t)
	id := uuid.New()

	mock.ExpectQuery(`UPDATE users SET email`).
		WithArgs("new@example.com", id).
		WillReturnError(sql.ErrConnDone)

	_, err := svc.UpdateEmail(context.Background(), UpdateEmailDTO{
		ID:    id,
		Email: "new@example.com",
	})
	if !errors.Is(err, ErrCouldNotUpdateEmail) {
		t.Errorf("UpdateEmail() error = %v, want %v", err, ErrCouldNotUpdateEmail)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestUpdatePassword_Success(t *testing.T) {
	svc, mock := newTestService(t)
	id := uuid.New()
	now := time.Now().UTC().Truncate(time.Microsecond)

	mock.ExpectQuery(`UPDATE users SET hashed_password`).
		WithArgs(sqlmock.AnyArg(), id).
		WillReturnRows(userRow(id, "user@example.com", now))

	res, err := svc.UpdatePassword(context.Background(), UpdatePasswordDTO{
		ID:       id,
		Password: "NewPass1!",
	})
	if err != nil {
		t.Fatalf("UpdatePassword() unexpected error: %v", err)
	}
	if res.ID != id {
		t.Errorf("res.ID = %s, want %s", res.ID, id)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestUpdatePassword_InvalidPassword(t *testing.T) {
	svc, _ := newTestService(t)

	_, err := svc.UpdatePassword(context.Background(), UpdatePasswordDTO{
		ID:       uuid.New(),
		Password: "weak",
	})
	if !errors.Is(err, password.ErrInvalidPassword) {
		t.Errorf("UpdatePassword() error = %v, want %v", err, password.ErrInvalidPassword)
	}
}

func TestUpdatePassword_DBFails(t *testing.T) {
	svc, mock := newTestService(t)
	id := uuid.New()

	mock.ExpectQuery(`UPDATE users SET hashed_password`).
		WithArgs(sqlmock.AnyArg(), id).
		WillReturnError(sql.ErrConnDone)

	_, err := svc.UpdatePassword(context.Background(), UpdatePasswordDTO{
		ID:       id,
		Password: "NewPass1!",
	})
	if !errors.Is(err, ErrCouldNotUpdatePass) {
		t.Errorf("UpdatePassword() error = %v, want %v", err, ErrCouldNotUpdatePass)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}