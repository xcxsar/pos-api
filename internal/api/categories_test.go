package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lib/pq"
	"github.com/xcxsar/pos-api/internal/category"
	"github.com/xcxsar/pos-api/internal/store/sqlc"
)

var categoryColumns = []string{"id", "name", "created_at", "updated_at"}

func newTestCategoryAPI(t *testing.T) (*API, sqlmock.Sqlmock, *http.ServeMux) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	svc := category.NewService(sqlc.New(db))
	app := NewAPI(nil, nil, nil, nil, svc)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/categories", app.CreateCategory)
	mux.HandleFunc("GET /api/categories", app.GetCategories)
	mux.HandleFunc("GET /api/categories/{categoryID}", app.GetCategoryByID)
	mux.HandleFunc("PUT /api/categories/{categoryID}", app.UpdateCategory)
	mux.HandleFunc("DELETE /api/categories/{categoryID}", app.DeleteCategory)

	return app, mock, mux
}

func categoryRow(id int64, name string, now time.Time) *sqlmock.Rows {
	return sqlmock.NewRows(categoryColumns).
		AddRow(id, name, now, now)
}

func assertCategoryBody(t *testing.T, rec *httptest.ResponseRecorder, wantName string) {
	t.Helper()
	var res category.Response
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("response is not valid JSON: %v (body: %s)", err, rec.Body.String())
	}
	if res.Name != wantName {
		t.Errorf("name = %q, want %q", res.Name, wantName)
	}
}

func TestCreateCategory_Success(t *testing.T) {
	_, mock, mux := newTestCategoryAPI(t)
	now := time.Now().UTC().Truncate(time.Microsecond)

	mock.ExpectQuery(`INSERT INTO categories`).
		WithArgs("Coffee").
		WillReturnRows(categoryRow(1, "Coffee", now))

	rec := doRequest(mux, http.MethodPost, "/api/categories", `{"name":"Coffee"}`)

	assertStatus(t, rec, http.StatusCreated)
	assertCategoryBody(t, rec, "Coffee")
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestCreateCategory_InvalidJSON(t *testing.T) {
	_, _, mux := newTestCategoryAPI(t)

	rec := doRequest(mux, http.MethodPost, "/api/categories", `{"name":"Coffee"`)

	assertStatus(t, rec, http.StatusBadRequest)
	assertErrorBody(t, rec, "invalid request format")
}

func TestCreateCategory_BlankNameDefaultsToUnnamed(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{"empty name", `{"name":""}`},
		{"whitespace-only name", `{"name":"   "}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, mock, mux := newTestCategoryAPI(t)
			now := time.Now().UTC().Truncate(time.Microsecond)

			mock.ExpectQuery(`INSERT INTO categories`).
				WithArgs("Unnamed").
				WillReturnRows(categoryRow(1, "Unnamed", now))

			rec := doRequest(mux, http.MethodPost, "/api/categories", tt.body)

			assertStatus(t, rec, http.StatusCreated)
			assertCategoryBody(t, rec, "Unnamed")
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unmet expectations: %v", err)
			}
		})
	}
}

func TestCreateCategory_TrimsName(t *testing.T) {
	_, mock, mux := newTestCategoryAPI(t)
	now := time.Now().UTC().Truncate(time.Microsecond)

	mock.ExpectQuery(`INSERT INTO categories`).
		WithArgs("Coffee").
		WillReturnRows(categoryRow(1, "Coffee", now))

	rec := doRequest(mux, http.MethodPost, "/api/categories", `{"name":"  Coffee  "}`)

	assertStatus(t, rec, http.StatusCreated)
	assertCategoryBody(t, rec, "Coffee")
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestCreateCategory_AcceptsUnicodeName(t *testing.T) {
	_, mock, mux := newTestCategoryAPI(t)
	now := time.Now().UTC().Truncate(time.Microsecond)

	mock.ExpectQuery(`INSERT INTO categories`).
		WithArgs("Café").
		WillReturnRows(categoryRow(1, "Café", now))

	rec := doRequest(mux, http.MethodPost, "/api/categories", `{"name":"Café"}`)

	assertStatus(t, rec, http.StatusCreated)
	assertCategoryBody(t, rec, "Café")
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestCreateCategory_InvalidCharacters(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{"digits", `{"name":"Coffee1"}`},
		{"internal space", `{"name":"Coffee Bean"}`},
		{"special characters", `{"name":"Coffee!"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, mux := newTestCategoryAPI(t)

			rec := doRequest(mux, http.MethodPost, "/api/categories", tt.body)

			assertStatus(t, rec, http.StatusBadRequest)
			assertErrorBody(t, rec, "category name cannot contain numbers, spaces or special characters")
		})
	}
}

func TestCreateCategory_DBError(t *testing.T) {
	_, mock, mux := newTestCategoryAPI(t)

	mock.ExpectQuery(`INSERT INTO categories`).
		WithArgs("Coffee").
		WillReturnError(sql.ErrConnDone)

	rec := doRequest(mux, http.MethodPost, "/api/categories", `{"name":"Coffee"}`)

	assertStatus(t, rec, http.StatusInternalServerError)
	assertErrorBody(t, rec, "could not process category registration")
}

func TestCreateCategory_UniqueViolation(t *testing.T) {
	_, mock, mux := newTestCategoryAPI(t)

	mock.ExpectQuery(`INSERT INTO categories`).
		WithArgs("Coffee").
		WillReturnError(&pq.Error{Code: "23505"})

	rec := doRequest(mux, http.MethodPost, "/api/categories", `{"name":"Coffee"}`)

	assertStatus(t, rec, http.StatusInternalServerError)
	assertErrorBody(t, rec, "could not process category registration")
}

func TestGetCategories_Success(t *testing.T) {
	_, mock, mux := newTestCategoryAPI(t)
	now := time.Now().UTC().Truncate(time.Microsecond)

	rows := sqlmock.NewRows(categoryColumns).
		AddRow(1, "Coffee", now, now).
		AddRow(2, "Tea", now, now)
	mock.ExpectQuery(`SELECT .+ FROM categories`).WillReturnRows(rows)

	rec := doRequest(mux, http.MethodGet, "/api/categories", "")

	assertStatus(t, rec, http.StatusOK)
	var res []category.Response
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("response is not valid JSON: %v (body: %s)", err, rec.Body.String())
	}
	if len(res) != 2 {
		t.Fatalf("returned %d categories, want 2", len(res))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestGetCategories_Empty(t *testing.T) {
	_, mock, mux := newTestCategoryAPI(t)

	mock.ExpectQuery(`SELECT .+ FROM categories`).WillReturnRows(sqlmock.NewRows(categoryColumns))

	rec := doRequest(mux, http.MethodGet, "/api/categories", "")

	assertStatus(t, rec, http.StatusOK)
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestGetCategories_DBError(t *testing.T) {
	_, mock, mux := newTestCategoryAPI(t)

	mock.ExpectQuery(`SELECT .+ FROM categories`).WillReturnError(sql.ErrConnDone)

	rec := doRequest(mux, http.MethodGet, "/api/categories", "")

	assertStatus(t, rec, http.StatusInternalServerError)
	assertErrorBody(t, rec, "could not retrieve categories")
}

func TestGetCategoryByID_Success(t *testing.T) {
	_, mock, mux := newTestCategoryAPI(t)
	now := time.Now().UTC().Truncate(time.Microsecond)

	mock.ExpectQuery(`SELECT .+ FROM categories WHERE id`).
		WithArgs(int64(7)).
		WillReturnRows(categoryRow(7, "Coffee", now))

	rec := doRequest(mux, http.MethodGet, "/api/categories/7", "")

	assertStatus(t, rec, http.StatusOK)
	assertCategoryBody(t, rec, "Coffee")
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestGetCategoryByID_InvalidID(t *testing.T) {
	_, _, mux := newTestCategoryAPI(t)

	rec := doRequest(mux, http.MethodGet, "/api/categories/abc", "")

	assertStatus(t, rec, http.StatusBadRequest)
	assertErrorBody(t, rec, "invalid category ID")
}

func TestGetCategoryByID_OverflowID(t *testing.T) {
	_, _, mux := newTestCategoryAPI(t)

	rec := doRequest(mux, http.MethodGet, "/api/categories/999999999999999999999999", "")

	assertStatus(t, rec, http.StatusBadRequest)
	assertErrorBody(t, rec, "invalid category ID")
}

func TestGetCategoryByID_NotFound(t *testing.T) {
	_, mock, mux := newTestCategoryAPI(t)

	mock.ExpectQuery(`SELECT .+ FROM categories WHERE id`).
		WithArgs(int64(7)).
		WillReturnError(sql.ErrNoRows)

	rec := doRequest(mux, http.MethodGet, "/api/categories/7", "")

	assertStatus(t, rec, http.StatusNotFound)
	assertErrorBody(t, rec, "requested category does not exist")
}

func TestGetCategoryByID_DBError(t *testing.T) {
	_, mock, mux := newTestCategoryAPI(t)

	mock.ExpectQuery(`SELECT .+ FROM categories WHERE id`).
		WithArgs(int64(7)).
		WillReturnError(sql.ErrConnDone)

	rec := doRequest(mux, http.MethodGet, "/api/categories/7", "")

	assertStatus(t, rec, http.StatusInternalServerError)
	assertErrorBody(t, rec, "could not retrieve category")
}

func TestUpdateCategory_Success(t *testing.T) {
	_, mock, mux := newTestCategoryAPI(t)
	now := time.Now().UTC().Truncate(time.Microsecond)

	mock.ExpectQuery(`SELECT .+ FROM categories WHERE id`).
		WithArgs(int64(5)).
		WillReturnRows(categoryRow(5, "Coffee", now))
	mock.ExpectQuery(`UPDATE categories SET name`).
		WithArgs("Milk", int64(5)).
		WillReturnRows(categoryRow(5, "Milk", now))

	rec := doRequest(mux, http.MethodPut, "/api/categories/5", `{"name":"Milk"}`)

	assertStatus(t, rec, http.StatusOK)
	assertCategoryBody(t, rec, "Milk")
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestUpdateCategory_InvalidID(t *testing.T) {
	_, _, mux := newTestCategoryAPI(t)

	rec := doRequest(mux, http.MethodPut, "/api/categories/abc", `{"name":"Milk"}`)

	assertStatus(t, rec, http.StatusBadRequest)
	assertErrorBody(t, rec, "invalid category ID")
}

func TestUpdateCategory_NotFound(t *testing.T) {
	_, mock, mux := newTestCategoryAPI(t)

	mock.ExpectQuery(`SELECT .+ FROM categories WHERE id`).
		WithArgs(int64(999)).
		WillReturnError(sql.ErrNoRows)

	rec := doRequest(mux, http.MethodPut, "/api/categories/999", `{"name":"Milk"}`)

	assertStatus(t, rec, http.StatusNotFound)
	assertErrorBody(t, rec, "targeted category does not exist")
}

func TestUpdateCategory_PreCheckDBError(t *testing.T) {
	_, mock, mux := newTestCategoryAPI(t)

	mock.ExpectQuery(`SELECT .+ FROM categories WHERE id`).
		WithArgs(int64(5)).
		WillReturnError(sql.ErrConnDone)

	rec := doRequest(mux, http.MethodPut, "/api/categories/5", `{"name":"Milk"}`)

	assertStatus(t, rec, http.StatusInternalServerError)
	assertErrorBody(t, rec, "could not retrieve category")
}

func TestUpdateCategory_InvalidJSON(t *testing.T) {
	_, mock, mux := newTestCategoryAPI(t)
	now := time.Now().UTC().Truncate(time.Microsecond)

	mock.ExpectQuery(`SELECT .+ FROM categories WHERE id`).
		WithArgs(int64(5)).
		WillReturnRows(categoryRow(5, "Coffee", now))

	rec := doRequest(mux, http.MethodPut, "/api/categories/5", `{"name":"Milk"`)

	assertStatus(t, rec, http.StatusBadRequest)
	assertErrorBody(t, rec, "invalid request format")
}

func TestUpdateCategory_BlankNameDefaultsToUnnamed(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{"empty name", `{"name":""}`},
		{"whitespace-only name", `{"name":"   "}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, mock, mux := newTestCategoryAPI(t)
			now := time.Now().UTC().Truncate(time.Microsecond)

			mock.ExpectQuery(`SELECT .+ FROM categories WHERE id`).
				WithArgs(int64(5)).
				WillReturnRows(categoryRow(5, "Coffee", now))
			mock.ExpectQuery(`UPDATE categories SET name`).
				WithArgs("Unnamed", int64(5)).
				WillReturnRows(categoryRow(5, "Unnamed", now))

			rec := doRequest(mux, http.MethodPut, "/api/categories/5", tt.body)

			assertStatus(t, rec, http.StatusOK)
			assertCategoryBody(t, rec, "Unnamed")
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unmet expectations: %v", err)
			}
		})
	}
}

func TestUpdateCategory_TrimsName(t *testing.T) {
	_, mock, mux := newTestCategoryAPI(t)
	now := time.Now().UTC().Truncate(time.Microsecond)

	mock.ExpectQuery(`SELECT .+ FROM categories WHERE id`).
		WithArgs(int64(5)).
		WillReturnRows(categoryRow(5, "Coffee", now))
	mock.ExpectQuery(`UPDATE categories SET name`).
		WithArgs("Milk", int64(5)).
		WillReturnRows(categoryRow(5, "Milk", now))

	rec := doRequest(mux, http.MethodPut, "/api/categories/5", `{"name":"  Milk  "}`)

	assertStatus(t, rec, http.StatusOK)
	assertCategoryBody(t, rec, "Milk")
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestUpdateCategory_InvalidCharacters(t *testing.T) {
	_, mock, mux := newTestCategoryAPI(t)
	now := time.Now().UTC().Truncate(time.Microsecond)

	mock.ExpectQuery(`SELECT .+ FROM categories WHERE id`).
		WithArgs(int64(5)).
		WillReturnRows(categoryRow(5, "Coffee", now))

	rec := doRequest(mux, http.MethodPut, "/api/categories/5", `{"name":"Milk!"}`)

	assertStatus(t, rec, http.StatusBadRequest)
	assertErrorBody(t, rec, "category name cannot contain numbers, spaces or special characters")
}

func TestUpdateCategory_DBError(t *testing.T) {
	_, mock, mux := newTestCategoryAPI(t)
	now := time.Now().UTC().Truncate(time.Microsecond)

	mock.ExpectQuery(`SELECT .+ FROM categories WHERE id`).
		WithArgs(int64(5)).
		WillReturnRows(categoryRow(5, "Coffee", now))
	mock.ExpectQuery(`UPDATE categories SET name`).
		WithArgs("Milk", int64(5)).
		WillReturnError(sql.ErrConnDone)

	rec := doRequest(mux, http.MethodPut, "/api/categories/5", `{"name":"Milk"}`)

	assertStatus(t, rec, http.StatusInternalServerError)
	assertErrorBody(t, rec, "failed updating infrastructure records")
}

func TestUpdateCategory_UniqueViolation(t *testing.T) {
	_, mock, mux := newTestCategoryAPI(t)
	now := time.Now().UTC().Truncate(time.Microsecond)

	mock.ExpectQuery(`SELECT .+ FROM categories WHERE id`).
		WithArgs(int64(5)).
		WillReturnRows(categoryRow(5, "Coffee", now))
	mock.ExpectQuery(`UPDATE categories SET name`).
		WithArgs("Milk", int64(5)).
		WillReturnError(&pq.Error{Code: "23505"})

	rec := doRequest(mux, http.MethodPut, "/api/categories/5", `{"name":"Milk"}`)

	assertStatus(t, rec, http.StatusInternalServerError)
	assertErrorBody(t, rec, "failed updating infrastructure records")
}

func TestDeleteCategory_Success(t *testing.T) {
	_, mock, mux := newTestCategoryAPI(t)

	mock.ExpectExec(`DELETE FROM categories`).
		WithArgs(int64(3)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	rec := doRequest(mux, http.MethodDelete, "/api/categories/3", "")

	assertStatus(t, rec, http.StatusNoContent)
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestDeleteCategory_InvalidID(t *testing.T) {
	_, _, mux := newTestCategoryAPI(t)

	rec := doRequest(mux, http.MethodDelete, "/api/categories/abc", "")

	assertStatus(t, rec, http.StatusBadRequest)
	assertErrorBody(t, rec, "invalid target identifier formatting")
}

func TestDeleteCategory_DBError(t *testing.T) {
	_, mock, mux := newTestCategoryAPI(t)

	mock.ExpectExec(`DELETE FROM categories`).
		WithArgs(int64(3)).
		WillReturnError(sql.ErrConnDone)

	rec := doRequest(mux, http.MethodDelete, "/api/categories/3", "")

	assertStatus(t, rec, http.StatusInternalServerError)
	assertErrorBody(t, rec, "unable to wipe category entry database records")
}