package category

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lib/pq"
	"github.com/xcxsar/pos-api/internal/store/sqlc"
)

var categoryColumns = []string{"id", "name", "created_at", "updated_at"}

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

func categoryRow(id int64, name string, createdAt, updatedAt time.Time) *sqlmock.Rows {
	return sqlmock.NewRows(categoryColumns).
		AddRow(id, name, createdAt, updatedAt)
}

func TestValidateName(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want bool
	}{
		{"valid mixed case", "Coffee", true},
		{"valid lowercase", "coffee", true},
		{"valid uppercase", "COFFEE", true},
		{"reserved default name", "Unnamed", true},
		{"accented letters", "Café", true},
		{"non-latin letters", "Эспрессо", true},
		{"empty string", "", false},
		{"whitespace only", "   ", false},
		{"leading space", " Coffee", false},
		{"trailing space", "Coffee ", false},
		{"internal space", "Coffee Bean", false},
		{"digits", "Coffee1", false},
		{"special characters", "Coffee!", false},
		{"underscore", "Coffee_Beans", false},
		{"emoji", "Coffee☕", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := validateName(tt.in); got != tt.want {
				t.Errorf("validateName(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestCreate_Success(t *testing.T) {
	svc, mock := newTestService(t)
	now := time.Now().UTC().Truncate(time.Microsecond)

	mock.ExpectQuery(`INSERT INTO categories`).
		WithArgs("Coffee").
		WillReturnRows(categoryRow(1, "Coffee", now, now))

	res, err := svc.Create(context.Background(), CreateDTO{Name: "Coffee"})
	if err != nil {
		t.Fatalf("Create() unexpected error: %v", err)
	}
	if res.ID != 1 || res.Name != "Coffee" {
		t.Errorf("res = %+v, want ID 1 / Coffee", res)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestCreate_BlankNameDefaultsToUnnamed(t *testing.T) {
	tests := []struct {
		name string
		raw  string
	}{
		{"empty name", ""},
		{"whitespace-only name", "   "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, mock := newTestService(t)
			now := time.Now().UTC().Truncate(time.Microsecond)

			mock.ExpectQuery(`INSERT INTO categories`).
				WithArgs("Unnamed").
				WillReturnRows(categoryRow(1, "Unnamed", now, now))

			res, err := svc.Create(context.Background(), CreateDTO{Name: tt.raw})
			if err != nil {
				t.Fatalf("Create() unexpected error: %v", err)
			}
			if res.Name != "Unnamed" {
				t.Errorf("res.Name = %q, want Unnamed", res.Name)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unmet expectations: %v", err)
			}
		})
	}
}

func TestCreate_TrimsName(t *testing.T) {
	svc, mock := newTestService(t)
	now := time.Now().UTC().Truncate(time.Microsecond)

	mock.ExpectQuery(`INSERT INTO categories`).
		WithArgs("Coffee").
		WillReturnRows(categoryRow(1, "Coffee", now, now))

	res, err := svc.Create(context.Background(), CreateDTO{Name: "  Coffee  "})
	if err != nil {
		t.Fatalf("Create() unexpected error: %v", err)
	}
	if res.Name != "Coffee" {
		t.Errorf("res.Name = %q, want Coffee", res.Name)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestCreate_AcceptsUnnamedName(t *testing.T) {
	svc, mock := newTestService(t)
	now := time.Now().UTC().Truncate(time.Microsecond)

	mock.ExpectQuery(`INSERT INTO categories`).
		WithArgs("Unnamed").
		WillReturnRows(categoryRow(1, "Unnamed", now, now))

	res, err := svc.Create(context.Background(), CreateDTO{Name: "Unnamed"})
	if err != nil {
		t.Fatalf("Create() unexpected error for reserved name: %v", err)
	}
	if res.Name != "Unnamed" {
		t.Errorf("res.Name = %q, want Unnamed", res.Name)
	}
}

func TestCreate_AcceptsUnicodeName(t *testing.T) {
	svc, mock := newTestService(t)
	now := time.Now().UTC().Truncate(time.Microsecond)

	mock.ExpectQuery(`INSERT INTO categories`).
		WithArgs("Café").
		WillReturnRows(categoryRow(1, "Café", now, now))

	res, err := svc.Create(context.Background(), CreateDTO{Name: "Café"})
	if err != nil {
		t.Fatalf("Create() unexpected error for accented name: %v", err)
	}
	if res.Name != "Café" {
		t.Errorf("res.Name = %q, want Café", res.Name)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestCreate_InvalidCharacters(t *testing.T) {
	svc, _ := newTestService(t)

	tests := []struct {
		name string
		raw  string
	}{
		{"digits", "Coffee1"},
		{"internal space", "Coffee Bean"},
		{"special characters", "Coffee!"},
		{"emoji", "Coffee☕"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.Create(context.Background(), CreateDTO{Name: tt.raw})
			if !errors.Is(err, ErrInvalidCharacters) {
				t.Errorf("Create() error = %v, want ErrInvalidCharacters", err)
			}
		})
	}
}

func TestCreate_DBFails(t *testing.T) {
	svc, mock := newTestService(t)

	mock.ExpectQuery(`INSERT INTO categories`).
		WithArgs("Coffee").
		WillReturnError(sql.ErrConnDone)

	_, err := svc.Create(context.Background(), CreateDTO{Name: "Coffee"})
	if !errors.Is(err, sql.ErrConnDone) {
		t.Errorf("Create() error = %v, want sql.ErrConnDone", err)
	}
}

func TestCreate_UniqueViolation(t *testing.T) {
	svc, mock := newTestService(t)
	dupErr := &pq.Error{Code: "23505", Message: "duplicate key value violates unique constraint"}

	mock.ExpectQuery(`INSERT INTO categories`).
		WithArgs("Coffee").
		WillReturnError(dupErr)

	_, err := svc.Create(context.Background(), CreateDTO{Name: "Coffee"})
	if err != dupErr {
		t.Errorf("Create() error = %v, want duplicate key error to propagate", err)
	}
}

func TestList_Success(t *testing.T) {
	svc, mock := newTestService(t)
	now := time.Now().UTC().Truncate(time.Microsecond)

	rows := sqlmock.NewRows(categoryColumns).
		AddRow(1, "Coffee", now, now).
		AddRow(2, "Tea", now, now)
	mock.ExpectQuery(`SELECT .+ FROM categories`).WillReturnRows(rows)

	res, err := svc.List(context.Background())
	if err != nil {
		t.Fatalf("List() unexpected error: %v", err)
	}
	if len(res) != 2 {
		t.Fatalf("List() returned %d categories, want 2", len(res))
	}
	if res[0].Name != "Coffee" || res[1].Name != "Tea" {
		t.Errorf("res = %+v, want Coffee/Tea", res)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestList_Empty(t *testing.T) {
	svc, mock := newTestService(t)

	mock.ExpectQuery(`SELECT .+ FROM categories`).WillReturnRows(sqlmock.NewRows(categoryColumns))

	res, err := svc.List(context.Background())
	if err != nil {
		t.Fatalf("List() unexpected error: %v", err)
	}
	if len(res) != 0 {
		t.Errorf("List() returned %d categories, want 0", len(res))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestList_DBFails(t *testing.T) {
	svc, mock := newTestService(t)

	mock.ExpectQuery(`SELECT .+ FROM categories`).WillReturnError(sql.ErrConnDone)

	res, err := svc.List(context.Background())
	if !errors.Is(err, sql.ErrConnDone) {
		t.Errorf("List() error = %v, want sql.ErrConnDone", err)
	}
	if res != nil {
		t.Errorf("List() returned %v, want nil on error", res)
	}
}

func TestGetByID_Success(t *testing.T) {
	svc, mock := newTestService(t)
	now := time.Now().UTC().Truncate(time.Microsecond)

	mock.ExpectQuery(`SELECT .+ FROM categories WHERE id`).
		WithArgs(int64(42)).
		WillReturnRows(categoryRow(42, "Coffee", now, now))

	res, err := svc.GetByID(context.Background(), 42)
	if err != nil {
		t.Fatalf("GetByID() unexpected error: %v", err)
	}
	if res.ID != 42 || res.Name != "Coffee" {
		t.Errorf("res = %+v, want ID 42 / Coffee", res)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestGetByID_NotFound(t *testing.T) {
	svc, mock := newTestService(t)

	mock.ExpectQuery(`SELECT .+ FROM categories WHERE id`).
		WithArgs(int64(999)).
		WillReturnError(sql.ErrNoRows)

	_, err := svc.GetByID(context.Background(), 999)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("GetByID() error = %v, want sql.ErrNoRows", err)
	}
}

func TestUpdate_Success(t *testing.T) {
	svc, mock := newTestService(t)
	now := time.Now().UTC().Truncate(time.Microsecond)

	mock.ExpectQuery(`UPDATE categories SET name`).
		WithArgs("Milk", int64(5)).
		WillReturnRows(categoryRow(5, "Milk", now, now))

	res, err := svc.Update(context.Background(), UpdateDTO{ID: 5, Name: "Milk"})
	if err != nil {
		t.Fatalf("Update() unexpected error: %v", err)
	}
	if res.Name != "Milk" {
		t.Errorf("res.Name = %q, want Milk", res.Name)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestUpdate_BlankNameDefaultsToUnnamed(t *testing.T) {
	tests := []struct {
		name string
		raw  string
	}{
		{"empty name", ""},
		{"whitespace-only name", "   "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, mock := newTestService(t)
			now := time.Now().UTC().Truncate(time.Microsecond)

			mock.ExpectQuery(`UPDATE categories SET name`).
				WithArgs("Unnamed", int64(5)).
				WillReturnRows(categoryRow(5, "Unnamed", now, now))

			res, err := svc.Update(context.Background(), UpdateDTO{ID: 5, Name: tt.raw})
			if err != nil {
				t.Fatalf("Update() unexpected error: %v", err)
			}
			if res.Name != "Unnamed" {
				t.Errorf("res.Name = %q, want Unnamed", res.Name)
			}
		})
	}
}

func TestUpdate_TrimsName(t *testing.T) {
	svc, mock := newTestService(t)
	now := time.Now().UTC().Truncate(time.Microsecond)

	mock.ExpectQuery(`UPDATE categories SET name`).
		WithArgs("Coffee", int64(5)).
		WillReturnRows(categoryRow(5, "Coffee", now, now))

	res, err := svc.Update(context.Background(), UpdateDTO{ID: 5, Name: "  Coffee  "})
	if err != nil {
		t.Fatalf("Update() unexpected error: %v", err)
	}
	if res.Name != "Coffee" {
		t.Errorf("res.Name = %q, want Coffee", res.Name)
	}
}

func TestUpdate_AcceptsUnnamedName(t *testing.T) {
	svc, mock := newTestService(t)
	now := time.Now().UTC().Truncate(time.Microsecond)

	mock.ExpectQuery(`UPDATE categories SET name`).
		WithArgs("Unnamed", int64(5)).
		WillReturnRows(categoryRow(5, "Unnamed", now, now))

	res, err := svc.Update(context.Background(), UpdateDTO{ID: 5, Name: "Unnamed"})
	if err != nil {
		t.Fatalf("Update() unexpected error for reserved name: %v", err)
	}
	if res.Name != "Unnamed" {
		t.Errorf("res.Name = %q, want Unnamed", res.Name)
	}
}

func TestUpdate_InvalidCharacters(t *testing.T) {
	svc, _ := newTestService(t)

	tests := []struct {
		name string
		raw  string
	}{
		{"digits", "Coffee1"},
		{"internal space", "Coffee Bean"},
		{"special characters", "Coffee!"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.Update(context.Background(), UpdateDTO{ID: 5, Name: tt.raw})
			if !errors.Is(err, ErrInvalidCharacters) {
				t.Errorf("Update() error = %v, want ErrInvalidCharacters", err)
			}
		})
	}
}

func TestUpdate_NotFound(t *testing.T) {
	svc, mock := newTestService(t)

	mock.ExpectQuery(`UPDATE categories SET name`).
		WithArgs("Milk", int64(999)).
		WillReturnError(sql.ErrNoRows)

	_, err := svc.Update(context.Background(), UpdateDTO{ID: 999, Name: "Milk"})
	if !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("Update() error = %v, want sql.ErrNoRows", err)
	}
}

func TestUpdate_DBFails(t *testing.T) {
	svc, mock := newTestService(t)

	mock.ExpectQuery(`UPDATE categories SET name`).
		WithArgs("Milk", int64(5)).
		WillReturnError(sql.ErrConnDone)

	_, err := svc.Update(context.Background(), UpdateDTO{ID: 5, Name: "Milk"})
	if !errors.Is(err, sql.ErrConnDone) {
		t.Errorf("Update() error = %v, want sql.ErrConnDone", err)
	}
}

func TestUpdate_UniqueViolation(t *testing.T) {
	svc, mock := newTestService(t)
	dupErr := &pq.Error{Code: "23505", Message: "duplicate key value violates unique constraint"}

	mock.ExpectQuery(`UPDATE categories SET name`).
		WithArgs("Milk", int64(5)).
		WillReturnError(dupErr)

	_, err := svc.Update(context.Background(), UpdateDTO{ID: 5, Name: "Milk"})
	if err != dupErr {
		t.Errorf("Update() error = %v, want duplicate key error to propagate", err)
	}
}

func TestDelete_Success(t *testing.T) {
	svc, mock := newTestService(t)

	mock.ExpectExec(`DELETE FROM categories`).
		WithArgs(int64(9)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := svc.Delete(context.Background(), 9); err != nil {
		t.Errorf("Delete() unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestDelete_MissingRowReturnsNil(t *testing.T) {
	svc, mock := newTestService(t)

	mock.ExpectExec(`DELETE FROM categories`).
		WithArgs(int64(999)).
		WillReturnResult(sqlmock.NewResult(0, 0))

	if err := svc.Delete(context.Background(), 999); err != nil {
		t.Errorf("Delete() unexpected error for missing row: %v", err)
	}
}

func TestDelete_DBFails(t *testing.T) {
	svc, mock := newTestService(t)

	mock.ExpectExec(`DELETE FROM categories`).
		WithArgs(int64(9)).
		WillReturnError(sql.ErrConnDone)

	if err := svc.Delete(context.Background(), 9); !errors.Is(err, sql.ErrConnDone) {
		t.Errorf("Delete() error = %v, want sql.ErrConnDone", err)
	}
}