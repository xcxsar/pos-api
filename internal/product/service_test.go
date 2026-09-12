package product

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/shopspring/decimal"
	"github.com/xcxsar/pos-api/internal/store/sqlc"
)

var productColumns = []string{"id", "name", "price", "stock", "category_id", "created_at", "updated_at"}

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

func productRow(id int64, name, price string, stock int32, categoryID any, createdAt, updatedAt time.Time) *sqlmock.Rows {
	return sqlmock.NewRows(productColumns).
		AddRow(id, name, price, stock, categoryID, createdAt, updatedAt)
}

func TestCreate_SuccessWithoutCategory(t *testing.T) {
	svc, mock := newTestService(t)
	now := time.Now().UTC().Truncate(time.Microsecond)

	mock.ExpectQuery(`INSERT INTO products`).
		WithArgs("Espresso", "19.99", int32(50), sql.NullInt64{}).
		WillReturnRows(productRow(1, "Espresso", "19.99", int32(50), nil, now, now))

	res, err := svc.Create(context.Background(), CreateDTO{
		Name:  "Espresso",
		Price: decimal.RequireFromString("19.99"),
		Stock: 50,
	})
	if err != nil {
		t.Fatalf("Create() unexpected error: %v", err)
	}
	if res.ID != 1 {
		t.Errorf("res.ID = %d, want 1", res.ID)
	}
	if res.Name != "Espresso" {
		t.Errorf("res.Name = %q, want Espresso", res.Name)
	}
	if res.Price.String() != "19.99" {
		t.Errorf("res.Price = %s, want 19.99", res.Price.String())
	}
	if res.Stock != 50 {
		t.Errorf("res.Stock = %d, want 50", res.Stock)
	}
	if res.CategoryID != nil {
		t.Errorf("res.CategoryID = %v, want nil", *res.CategoryID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestCreate_WithCategory(t *testing.T) {
	svc, mock := newTestService(t)
	now := time.Now().UTC().Truncate(time.Microsecond)
	categoryID := int64(7)

	mock.ExpectQuery(`INSERT INTO products`).
		WithArgs("Espresso", "19.99", int32(50), sql.NullInt64{Int64: 7, Valid: true}).
		WillReturnRows(productRow(1, "Espresso", "19.99", int32(50), int64(7), now, now))

	res, err := svc.Create(context.Background(), CreateDTO{
		Name:       "Espresso",
		Price:      decimal.RequireFromString("19.99"),
		Stock:      50,
		CategoryID: &categoryID,
	})
	if err != nil {
		t.Fatalf("Create() unexpected error: %v", err)
	}
	if res.CategoryID == nil || *res.CategoryID != 7 {
		t.Errorf("res.CategoryID = %v, want ptr to 7", res.CategoryID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestCreate_Validation(t *testing.T) {
	svc, _ := newTestService(t)
	price := decimal.NewFromInt(10)

	tests := []struct {
		name string
		dto  CreateDTO
		want error
	}{
		{"negative price rejected", CreateDTO{Name: "Espresso", Price: decimal.NewFromInt(-1), Stock: 1}, ErrInvalidPrice},
		{"negative stock rejected", CreateDTO{Name: "Espresso", Price: price, Stock: -1}, ErrInvalidStock},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.Create(context.Background(), tt.dto)
			if !errors.Is(err, tt.want) {
				t.Errorf("Create() error = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestCreate_ZeroPriceAndStock(t *testing.T) {
	svc, mock := newTestService(t)
	now := time.Now().UTC().Truncate(time.Microsecond)

	mock.ExpectQuery(`INSERT INTO products`).
		WithArgs("Gratis", "0", int32(0), sql.NullInt64{}).
		WillReturnRows(productRow(1, "Gratis", "0", int32(0), nil, now, now))

	res, err := svc.Create(context.Background(), CreateDTO{
		Name:  "Gratis",
		Price: decimal.Zero,
		Stock: 0,
	})
	if err != nil {
		t.Fatalf("Create() unexpected error: %v", err)
	}
	if res.Price.String() != "0" || res.Stock != 0 {
		t.Errorf("res = %+v, want price 0 and stock 0", res)
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

			mock.ExpectQuery(`INSERT INTO products`).
				WithArgs("Unnamed", "10", int32(5), sql.NullInt64{}).
				WillReturnRows(productRow(1, "Unnamed", "10", int32(5), nil, now, now))

			res, err := svc.Create(context.Background(), CreateDTO{
				Name:  tt.raw,
				Price: decimal.RequireFromString("10.00"),
				Stock: 5,
			})
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

func TestCreate_AcceptsUnnamedName(t *testing.T) {
	svc, mock := newTestService(t)
	now := time.Now().UTC().Truncate(time.Microsecond)

	mock.ExpectQuery(`INSERT INTO products`).
		WithArgs("Unnamed", "10", int32(5), sql.NullInt64{}).
		WillReturnRows(productRow(1, "Unnamed", "10", int32(5), nil, now, now))

	res, err := svc.Create(context.Background(), CreateDTO{
		Name:  "Unnamed",
		Price: decimal.RequireFromString("10.00"),
		Stock: 5,
	})
	if err != nil {
		t.Fatalf("Create() unexpected error: %v", err)
	}
	if res.Name != "Unnamed" {
		t.Errorf("res.Name = %q, want Unnamed", res.Name)
	}
}

func TestCreate_TrimsName(t *testing.T) {
	svc, mock := newTestService(t)
	now := time.Now().UTC().Truncate(time.Microsecond)

	mock.ExpectQuery(`INSERT INTO products`).
		WithArgs("Espresso", "10", int32(5), sql.NullInt64{}).
		WillReturnRows(productRow(1, "Espresso", "10", int32(5), nil, now, now))

	res, err := svc.Create(context.Background(), CreateDTO{
		Name:  "  Espresso  ",
		Price: decimal.RequireFromString("10.00"),
		Stock: 5,
	})
	if err != nil {
		t.Fatalf("Create() unexpected error: %v", err)
	}
	if res.Name != "Espresso" {
		t.Errorf("res.Name = %q, want Espresso", res.Name)
	}
}

func TestCreate_DBFails(t *testing.T) {
	svc, mock := newTestService(t)

	mock.ExpectQuery(`INSERT INTO products`).
		WithArgs("Espresso", "19.99", int32(50), sql.NullInt64{}).
		WillReturnError(sql.ErrConnDone)

	_, err := svc.Create(context.Background(), CreateDTO{
		Name:  "Espresso",
		Price: decimal.RequireFromString("19.99"),
		Stock: 50,
	})
	if !errors.Is(err, sql.ErrConnDone) {
		t.Errorf("Create() error = %v, want sql.ErrConnDone", err)
	}
}

func TestCreate_InvalidPriceFromDB(t *testing.T) {
	svc, mock := newTestService(t)
	now := time.Now().UTC().Truncate(time.Microsecond)

	mock.ExpectQuery(`INSERT INTO products`).
		WithArgs("Weird", "999", int32(1), sql.NullInt64{}).
		WillReturnRows(productRow(1, "Weird", "not-a-number", int32(1), nil, now, now))

	_, err := svc.Create(context.Background(), CreateDTO{
		Name:  "Weird",
		Price: decimal.RequireFromString("999.00"),
		Stock: 1,
	})
	if err == nil {
		t.Fatal("Create() expected error for non-numeric price in DB row, got nil")
	}
	if !strings.Contains(err.Error(), "invalid price format") {
		t.Errorf("Create() error = %v, want invalid price format", err)
	}
}

func TestList_Success(t *testing.T) {
	svc, mock := newTestService(t)
	now := time.Now().UTC().Truncate(time.Microsecond)

	rows := sqlmock.NewRows(productColumns).
		AddRow(1, "Espresso", "10.5", int32(3), nil, now, now).
		AddRow(2, "Tea", "20", int32(5), int64(9), now, now)
	mock.ExpectQuery(`SELECT .+ FROM products`).WillReturnRows(rows)

	res, err := svc.List(context.Background())
	if err != nil {
		t.Fatalf("List() unexpected error: %v", err)
	}
	if len(res) != 2 {
		t.Fatalf("List() returned %d products, want 2", len(res))
	}
	if res[0].Name != "Espresso" || res[0].Price.String() != "10.5" {
		t.Errorf("res[0] = %+v, want Espresso/10.5", res[0])
	}
	if res[1].CategoryID == nil || *res[1].CategoryID != 9 {
		t.Errorf("res[1].CategoryID = %v, want ptr to 9", res[1].CategoryID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestList_Empty(t *testing.T) {
	svc, mock := newTestService(t)

	mock.ExpectQuery(`SELECT .+ FROM products`).WillReturnRows(sqlmock.NewRows(productColumns))

	res, err := svc.List(context.Background())
	if err != nil {
		t.Fatalf("List() unexpected error: %v", err)
	}
	if len(res) != 0 {
		t.Errorf("List() returned %d products, want 0", len(res))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestList_DBFails(t *testing.T) {
	svc, mock := newTestService(t)

	mock.ExpectQuery(`SELECT .+ FROM products`).WillReturnError(sql.ErrConnDone)

	res, err := svc.List(context.Background())
	if !errors.Is(err, sql.ErrConnDone) {
		t.Errorf("List() error = %v, want sql.ErrConnDone", err)
	}
	if res != nil {
		t.Errorf("List() returned %v, want nil on error", res)
	}
}

func TestList_InvalidPriceFromDB(t *testing.T) {
	svc, mock := newTestService(t)
	now := time.Now().UTC().Truncate(time.Microsecond)

	mock.ExpectQuery(`SELECT .+ FROM products`).
		WillReturnRows(productRow(1, "Bad", "oops", int32(2), nil, now, now))

	_, err := svc.List(context.Background())
	if err == nil {
		t.Fatal("List() expected error for non-numeric price, got nil")
	}
}

func TestGetByID_Success(t *testing.T) {
	svc, mock := newTestService(t)
	now := time.Now().UTC().Truncate(time.Microsecond)

	mock.ExpectQuery(`SELECT .+ FROM products WHERE id`).
		WithArgs(int64(42)).
		WillReturnRows(productRow(42, "Espresso", "19.99", int32(50), int64(7), now, now))

	res, err := svc.GetByID(context.Background(), 42)
	if err != nil {
		t.Fatalf("GetByID() unexpected error: %v", err)
	}
	if res.ID != 42 || res.Name != "Espresso" {
		t.Errorf("res = %+v, want ID 42 / Espresso", res)
	}
	if res.CategoryID == nil || *res.CategoryID != 7 {
		t.Errorf("res.CategoryID = %v, want ptr to 7", res.CategoryID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestGetByID_NullCategory(t *testing.T) {
	svc, mock := newTestService(t)
	now := time.Now().UTC().Truncate(time.Microsecond)

	mock.ExpectQuery(`SELECT .+ FROM products WHERE id`).
		WithArgs(int64(1)).
		WillReturnRows(productRow(1, "Espresso", "19.99", int32(50), nil, now, now))

	res, err := svc.GetByID(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetByID() unexpected error: %v", err)
	}
	if res.CategoryID != nil {
		t.Errorf("res.CategoryID = %v, want nil", *res.CategoryID)
	}
}

func TestGetByID_NotFound(t *testing.T) {
	svc, mock := newTestService(t)

	mock.ExpectQuery(`SELECT .+ FROM products WHERE id`).
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
	categoryID := int64(3)

	mock.ExpectQuery(`UPDATE products SET name`).
		WithArgs("Americano", "18.5", int32(20), sql.NullInt64{Int64: 3, Valid: true}, int64(5)).
		WillReturnRows(productRow(5, "Americano", "18.5", int32(20), int64(3), now, now))

	res, err := svc.Update(context.Background(), UpdateDTO{
		ID:         5,
		Name:       "Americano",
		Price:      decimal.RequireFromString("18.50"),
		Stock:      20,
		CategoryID: &categoryID,
	})
	if err != nil {
		t.Fatalf("Update() unexpected error: %v", err)
	}
	if res.Name != "Americano" || res.Price.String() != "18.5" || res.Stock != 20 {
		t.Errorf("res = %+v, want Americano/18.5/20", res)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestUpdate_AcceptsUnnamedName(t *testing.T) {
	svc, mock := newTestService(t)
	now := time.Now().UTC().Truncate(time.Microsecond)

	mock.ExpectQuery(`UPDATE products SET name`).
		WithArgs("Unnamed", "10", int32(1), sql.NullInt64{}, int64(5)).
		WillReturnRows(productRow(5, "Unnamed", "10", int32(1), nil, now, now))

	res, err := svc.Update(context.Background(), UpdateDTO{
		ID:    5,
		Name:  "Unnamed",
		Price: decimal.RequireFromString("10.00"),
		Stock: 1,
	})
	if err != nil {
		t.Fatalf("Update() unexpected error for reserved name: %v", err)
	}
	if res.Name != "Unnamed" {
		t.Errorf("res.Name = %q, want Unnamed", res.Name)
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

			mock.ExpectQuery(`UPDATE products SET name`).
				WithArgs("Unnamed", "10", int32(1), sql.NullInt64{}, int64(5)).
				WillReturnRows(productRow(5, "Unnamed", "10", int32(1), nil, now, now))

			res, err := svc.Update(context.Background(), UpdateDTO{
				ID:    5,
				Name:  tt.raw,
				Price: decimal.RequireFromString("10.00"),
				Stock: 1,
			})
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

	mock.ExpectQuery(`UPDATE products SET name`).
		WithArgs("Espresso", "10", int32(1), sql.NullInt64{}, int64(5)).
		WillReturnRows(productRow(5, "Espresso", "10", int32(1), nil, now, now))

	res, err := svc.Update(context.Background(), UpdateDTO{
		ID:    5,
		Name:  "  Espresso  ",
		Price: decimal.RequireFromString("10.00"),
		Stock: 1,
	})
	if err != nil {
		t.Fatalf("Update() unexpected error: %v", err)
	}
	if res.Name != "Espresso" {
		t.Errorf("res.Name = %q, want Espresso", res.Name)
	}
}

func TestUpdate_Validation(t *testing.T) {
	svc, _ := newTestService(t)
	price := decimal.NewFromInt(10)

	tests := []struct {
		name string
		dto  UpdateDTO
		want error
	}{
		{"negative price rejected", UpdateDTO{ID: 1, Name: "Espresso", Price: decimal.NewFromInt(-1), Stock: 1}, ErrInvalidPrice},
		{"negative stock rejected", UpdateDTO{ID: 1, Name: "Espresso", Price: price, Stock: -1}, ErrInvalidStock},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.Update(context.Background(), tt.dto)
			if !errors.Is(err, tt.want) {
				t.Errorf("Update() error = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestUpdate_NotFound(t *testing.T) {
	svc, mock := newTestService(t)

	mock.ExpectQuery(`UPDATE products SET name`).
		WithArgs("Americano", "18.5", int32(20), sql.NullInt64{}, int64(999)).
		WillReturnError(sql.ErrNoRows)

	_, err := svc.Update(context.Background(), UpdateDTO{
		ID:    999,
		Name:  "Americano",
		Price: decimal.RequireFromString("18.50"),
		Stock: 20,
	})
	if !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("Update() error = %v, want sql.ErrNoRows", err)
	}
}

func TestUpdate_DBFails(t *testing.T) {
	svc, mock := newTestService(t)

	mock.ExpectQuery(`UPDATE products SET name`).
		WithArgs("Americano", "18.5", int32(20), sql.NullInt64{}, int64(5)).
		WillReturnError(sql.ErrConnDone)

	_, err := svc.Update(context.Background(), UpdateDTO{
		ID:    5,
		Name:  "Americano",
		Price: decimal.RequireFromString("18.50"),
		Stock: 20,
	})
	if !errors.Is(err, sql.ErrConnDone) {
		t.Errorf("Update() error = %v, want sql.ErrConnDone", err)
	}
}

func TestDelete_Success(t *testing.T) {
	svc, mock := newTestService(t)

	mock.ExpectExec(`DELETE FROM products`).
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

	mock.ExpectExec(`DELETE FROM products`).
		WithArgs(int64(999)).
		WillReturnResult(sqlmock.NewResult(0, 0))

	if err := svc.Delete(context.Background(), 999); err != nil {
		t.Errorf("Delete() unexpected error for missing row: %v", err)
	}
}

func TestDelete_DBFails(t *testing.T) {
	svc, mock := newTestService(t)

	mock.ExpectExec(`DELETE FROM products`).
		WithArgs(int64(9)).
		WillReturnError(sql.ErrConnDone)

	if err := svc.Delete(context.Background(), 9); !errors.Is(err, sql.ErrConnDone) {
		t.Errorf("Delete() error = %v, want sql.ErrConnDone", err)
	}
}