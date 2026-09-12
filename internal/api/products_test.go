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
	"github.com/xcxsar/pos-api/internal/product"
	"github.com/xcxsar/pos-api/internal/store/sqlc"
)

var productColumns = []string{"id", "name", "price", "stock", "category_id", "created_at", "updated_at"}

func newTestAPI(t *testing.T) (*API, sqlmock.Sqlmock, *http.ServeMux) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	svc := product.NewService(sqlc.New(db))
	app := NewAPI(nil, svc, nil, nil, nil)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/products", app.CreateProduct)
	mux.HandleFunc("GET /api/products", app.GetProducts)
	mux.HandleFunc("GET /api/products/{productID}", app.GetProductByID)
	mux.HandleFunc("PUT /api/products/{productID}", app.UpdateProduct)
	mux.HandleFunc("DELETE /api/products/{productID}", app.DeleteProduct)

	return app, mock, mux
}

func productRow(id int64, name, price string, stock int32, categoryID any, now time.Time) *sqlmock.Rows {
	return sqlmock.NewRows(productColumns).
		AddRow(id, name, price, stock, categoryID, now, now)
}

func doRequest(mux *http.ServeMux, method, path, body string) *httptest.ResponseRecorder {
	var bodyReader io.Reader
	if body != "" {
		bodyReader = bytes.NewBufferString(body)
	}
	req := httptest.NewRequest(method, path, bodyReader)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	return rec
}

func assertStatus(t *testing.T, rec *httptest.ResponseRecorder, want int) {
	t.Helper()
	if rec.Code != want {
		t.Fatalf("status = %d, want %d (body: %s)", rec.Code, want, rec.Body.String())
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}
}

func assertErrorBody(t *testing.T, rec *httptest.ResponseRecorder, wantMsg string) {
	t.Helper()
	var errRes ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &errRes); err != nil {
		t.Fatalf("response is not valid JSON: %v (body: %s)", err, rec.Body.String())
	}
	if errRes.Error != wantMsg {
		t.Errorf("error message = %q, want %q", errRes.Error, wantMsg)
	}
}

func assertProductBody(t *testing.T, rec *httptest.ResponseRecorder, wantName, wantPrice string) {
	t.Helper()
	var res product.Response
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("response is not valid JSON: %v (body: %s)", err, rec.Body.String())
	}
	if res.Name != wantName {
		t.Errorf("name = %q, want %q", res.Name, wantName)
	}
	if res.Price.String() != wantPrice {
		t.Errorf("price = %s, want %s", res.Price.String(), wantPrice)
	}
}

func TestCreateProduct_Success(t *testing.T) {
	_, mock, mux := newTestAPI(t)
	now := time.Now().UTC().Truncate(time.Microsecond)

	mock.ExpectQuery(`INSERT INTO products`).
		WithArgs("Espresso", "19.99", int32(50), sql.NullInt64{}).
		WillReturnRows(productRow(1, "Espresso", "19.99", int32(50), nil, now))

	rec := doRequest(mux, http.MethodPost, "/api/products", `{"name":"Espresso","price":19.99,"stock":50}`)

	assertStatus(t, rec, http.StatusCreated)
	assertProductBody(t, rec, "Espresso", "19.99")
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestCreateProduct_WithCategory(t *testing.T) {
	_, mock, mux := newTestAPI(t)
	now := time.Now().UTC().Truncate(time.Microsecond)

	mock.ExpectQuery(`INSERT INTO products`).
		WithArgs("Espresso", "19.99", int32(50), sql.NullInt64{Int64: 5, Valid: true}).
		WillReturnRows(productRow(1, "Espresso", "19.99", int32(50), int64(5), now))

	rec := doRequest(mux, http.MethodPost, "/api/products", `{"name":"Espresso","price":19.99,"stock":50,"category_id":5}`)

	assertStatus(t, rec, http.StatusCreated)
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestCreateProduct_InvalidJSON(t *testing.T) {
	_, _, mux := newTestAPI(t)

	rec := doRequest(mux, http.MethodPost, "/api/products", `{"name":"Espresso"`)

	assertStatus(t, rec, http.StatusBadRequest)
	assertErrorBody(t, rec, "invalid request format")
}

func TestCreateProduct_ValidationError(t *testing.T) {
	tests := []struct {
		name string
		body string
		want string
	}{
		{"negative price", `{"name":"Espresso","price":-1,"stock":1}`, "product price cannot be negative"},
		{"negative stock", `{"name":"Espresso","price":1,"stock":-1}`, "product stock cannot be negative"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, mux := newTestAPI(t)

			rec := doRequest(mux, http.MethodPost, "/api/products", tt.body)

			assertStatus(t, rec, http.StatusBadRequest)
			assertErrorBody(t, rec, tt.want)
		})
	}
}

func TestCreateProduct_BlankNameDefaultsToUnnamed(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{"empty name", `{"name":"","price":19.99,"stock":50}`},
		{"whitespace-only name", `{"name":"   ","price":19.99,"stock":50}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, mock, mux := newTestAPI(t)
			now := time.Now().UTC().Truncate(time.Microsecond)

			mock.ExpectQuery(`INSERT INTO products`).
				WithArgs("Unnamed", "19.99", int32(50), sql.NullInt64{}).
				WillReturnRows(productRow(1, "Unnamed", "19.99", int32(50), nil, now))

			rec := doRequest(mux, http.MethodPost, "/api/products", tt.body)

			assertStatus(t, rec, http.StatusCreated)
			assertProductBody(t, rec, "Unnamed", "19.99")
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unmet expectations: %v", err)
			}
		})
	}
}

func TestCreateProduct_TrimsName(t *testing.T) {
	_, mock, mux := newTestAPI(t)
	now := time.Now().UTC().Truncate(time.Microsecond)

	mock.ExpectQuery(`INSERT INTO products`).
		WithArgs("Espresso", "19.99", int32(50), sql.NullInt64{}).
		WillReturnRows(productRow(1, "Espresso", "19.99", int32(50), nil, now))

	rec := doRequest(mux, http.MethodPost, "/api/products", `{"name":"  Espresso  ","price":19.99,"stock":50}`)

	assertStatus(t, rec, http.StatusCreated)
	assertProductBody(t, rec, "Espresso", "19.99")
}

func TestCreateProduct_DBError(t *testing.T) {
	_, mock, mux := newTestAPI(t)

	mock.ExpectQuery(`INSERT INTO products`).
		WithArgs("Espresso", "19.99", int32(50), sql.NullInt64{}).
		WillReturnError(sql.ErrConnDone)

	rec := doRequest(mux, http.MethodPost, "/api/products", `{"name":"Espresso","price":19.99,"stock":50}`)

	assertStatus(t, rec, http.StatusInternalServerError)
	assertErrorBody(t, rec, "could not process product registration")
}

func TestGetProducts_Success(t *testing.T) {
	_, mock, mux := newTestAPI(t)
	now := time.Now().UTC().Truncate(time.Microsecond)

	rows := sqlmock.NewRows(productColumns).
		AddRow(1, "Espresso", "10.5", int32(3), nil, now, now).
		AddRow(2, "Tea", "20", int32(5), int64(9), now, now)
	mock.ExpectQuery(`SELECT .+ FROM products`).WillReturnRows(rows)

	rec := doRequest(mux, http.MethodGet, "/api/products", "")

	assertStatus(t, rec, http.StatusOK)
	var res []product.Response
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("response is not valid JSON: %v (body: %s)", err, rec.Body.String())
	}
	if len(res) != 2 {
		t.Fatalf("returned %d products, want 2", len(res))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestGetProducts_Empty(t *testing.T) {
	_, mock, mux := newTestAPI(t)

	mock.ExpectQuery(`SELECT .+ FROM products`).WillReturnRows(sqlmock.NewRows(productColumns))

	rec := doRequest(mux, http.MethodGet, "/api/products", "")

	assertStatus(t, rec, http.StatusOK)
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestGetProducts_DBError(t *testing.T) {
	_, mock, mux := newTestAPI(t)

	mock.ExpectQuery(`SELECT .+ FROM products`).WillReturnError(sql.ErrConnDone)

	rec := doRequest(mux, http.MethodGet, "/api/products", "")

	assertStatus(t, rec, http.StatusInternalServerError)
	assertErrorBody(t, rec, "could not retrieve products")
}

func TestGetProductByID_Success(t *testing.T) {
	_, mock, mux := newTestAPI(t)
	now := time.Now().UTC().Truncate(time.Microsecond)

	mock.ExpectQuery(`SELECT .+ FROM products WHERE id`).
		WithArgs(int64(7)).
		WillReturnRows(productRow(7, "Espresso", "19.99", int32(50), nil, now))

	rec := doRequest(mux, http.MethodGet, "/api/products/7", "")

	assertStatus(t, rec, http.StatusOK)
	assertProductBody(t, rec, "Espresso", "19.99")
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestGetProductByID_InvalidID(t *testing.T) {
	_, _, mux := newTestAPI(t)

	rec := doRequest(mux, http.MethodGet, "/api/products/abc", "")

	assertStatus(t, rec, http.StatusBadRequest)
	assertErrorBody(t, rec, "invalid product ID")
}

func TestGetProductByID_OverflowID(t *testing.T) {
	_, _, mux := newTestAPI(t)

	rec := doRequest(mux, http.MethodGet, "/api/products/999999999999999999999999", "")

	assertStatus(t, rec, http.StatusBadRequest)
	assertErrorBody(t, rec, "invalid product ID")
}

func TestGetProductByID_NotFound(t *testing.T) {
	_, mock, mux := newTestAPI(t)

	mock.ExpectQuery(`SELECT .+ FROM products WHERE id`).
		WithArgs(int64(7)).
		WillReturnError(sql.ErrNoRows)

	rec := doRequest(mux, http.MethodGet, "/api/products/7", "")

	assertStatus(t, rec, http.StatusNotFound)
	assertErrorBody(t, rec, "requested product does not exist")
}

func TestGetProductByID_DBError(t *testing.T) {
	_, mock, mux := newTestAPI(t)

	mock.ExpectQuery(`SELECT .+ FROM products WHERE id`).
		WithArgs(int64(7)).
		WillReturnError(sql.ErrConnDone)

	rec := doRequest(mux, http.MethodGet, "/api/products/7", "")

	assertStatus(t, rec, http.StatusInternalServerError)
	assertErrorBody(t, rec, "could not retrieve product")
}

func TestUpdateProduct_Success(t *testing.T) {
	_, mock, mux := newTestAPI(t)
	now := time.Now().UTC().Truncate(time.Microsecond)

	mock.ExpectQuery(`SELECT .+ FROM products WHERE id`).
		WithArgs(int64(5)).
		WillReturnRows(productRow(5, "Espresso", "19.99", int32(50), nil, now))
	mock.ExpectQuery(`UPDATE products SET name`).
		WithArgs("Americano", "18.5", int32(20), sql.NullInt64{}, int64(5)).
		WillReturnRows(productRow(5, "Americano", "18.5", int32(20), nil, now))

	rec := doRequest(mux, http.MethodPut, "/api/products/5", `{"name":"Americano","price":18.5,"stock":20}`)

	assertStatus(t, rec, http.StatusOK)
	assertProductBody(t, rec, "Americano", "18.5")
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestUpdateProduct_InvalidID(t *testing.T) {
	_, _, mux := newTestAPI(t)

	rec := doRequest(mux, http.MethodPut, "/api/products/abc", `{"name":"Americano","price":1,"stock":1}`)

	assertStatus(t, rec, http.StatusBadRequest)
	assertErrorBody(t, rec, "invalid product ID")
}

func TestUpdateProduct_NotFound(t *testing.T) {
	_, mock, mux := newTestAPI(t)

	mock.ExpectQuery(`SELECT .+ FROM products WHERE id`).
		WithArgs(int64(999)).
		WillReturnError(sql.ErrNoRows)

	rec := doRequest(mux, http.MethodPut, "/api/products/999", `{"name":"Americano","price":1,"stock":1}`)

	assertStatus(t, rec, http.StatusNotFound)
	assertErrorBody(t, rec, "targeted product does not exist")
}

func TestUpdateProduct_PreCheckDBError(t *testing.T) {
	_, mock, mux := newTestAPI(t)

	mock.ExpectQuery(`SELECT .+ FROM products WHERE id`).
		WithArgs(int64(5)).
		WillReturnError(sql.ErrConnDone)

	rec := doRequest(mux, http.MethodPut, "/api/products/5", `{"name":"Americano","price":1,"stock":1}`)

	assertStatus(t, rec, http.StatusInternalServerError)
	assertErrorBody(t, rec, "could not retrieve product")
}

func TestUpdateProduct_InvalidJSON(t *testing.T) {
	_, mock, mux := newTestAPI(t)
	now := time.Now().UTC().Truncate(time.Microsecond)

	mock.ExpectQuery(`SELECT .+ FROM products WHERE id`).
		WithArgs(int64(5)).
		WillReturnRows(productRow(5, "Espresso", "19.99", int32(50), nil, now))

	rec := doRequest(mux, http.MethodPut, "/api/products/5", `{"name":"Americano"`)

	assertStatus(t, rec, http.StatusBadRequest)
	assertErrorBody(t, rec, "invalid parameters format")
}

func TestUpdateProduct_ValidationError(t *testing.T) {
	tests := []struct {
		name string
		body string
		want string
	}{
		{"negative price", `{"name":"Americano","price":-1,"stock":1}`, "product price cannot be negative"},
		{"negative stock", `{"name":"Americano","price":1,"stock":-1}`, "product stock cannot be negative"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, mock, mux := newTestAPI(t)
			now := time.Now().UTC().Truncate(time.Microsecond)

			mock.ExpectQuery(`SELECT .+ FROM products WHERE id`).
				WithArgs(int64(5)).
				WillReturnRows(productRow(5, "Espresso", "19.99", int32(50), nil, now))

			rec := doRequest(mux, http.MethodPut, "/api/products/5", tt.body)

			assertStatus(t, rec, http.StatusBadRequest)
			assertErrorBody(t, rec, tt.want)
		})
	}
}

func TestUpdateProduct_BlankNameDefaultsToUnnamed(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{"empty name", `{"name":"","price":18.5,"stock":20}`},
		{"whitespace-only name", `{"name":"   ","price":18.5,"stock":20}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, mock, mux := newTestAPI(t)
			now := time.Now().UTC().Truncate(time.Microsecond)

			mock.ExpectQuery(`SELECT .+ FROM products WHERE id`).
				WithArgs(int64(5)).
				WillReturnRows(productRow(5, "Espresso", "19.99", int32(50), nil, now))
			mock.ExpectQuery(`UPDATE products SET name`).
				WithArgs("Unnamed", "18.5", int32(20), sql.NullInt64{}, int64(5)).
				WillReturnRows(productRow(5, "Unnamed", "18.5", int32(20), nil, now))

			rec := doRequest(mux, http.MethodPut, "/api/products/5", tt.body)

			assertStatus(t, rec, http.StatusOK)
			assertProductBody(t, rec, "Unnamed", "18.5")
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unmet expectations: %v", err)
			}
		})
	}
}

func TestUpdateProduct_DBError(t *testing.T) {
	_, mock, mux := newTestAPI(t)
	now := time.Now().UTC().Truncate(time.Microsecond)

	mock.ExpectQuery(`SELECT .+ FROM products WHERE id`).
		WithArgs(int64(5)).
		WillReturnRows(productRow(5, "Espresso", "19.99", int32(50), nil, now))
	mock.ExpectQuery(`UPDATE products SET name`).
		WithArgs("Americano", "18.5", int32(20), sql.NullInt64{}, int64(5)).
		WillReturnError(sql.ErrConnDone)

	rec := doRequest(mux, http.MethodPut, "/api/products/5", `{"name":"Americano","price":18.5,"stock":20}`)

	assertStatus(t, rec, http.StatusInternalServerError)
	assertErrorBody(t, rec, "failed updating infrastructure records")
}

func TestDeleteProduct_Success(t *testing.T) {
	_, mock, mux := newTestAPI(t)

	mock.ExpectExec(`DELETE FROM products`).
		WithArgs(int64(3)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	rec := doRequest(mux, http.MethodDelete, "/api/products/3", "")

	assertStatus(t, rec, http.StatusNoContent)
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestDeleteProduct_InvalidID(t *testing.T) {
	_, _, mux := newTestAPI(t)

	rec := doRequest(mux, http.MethodDelete, "/api/products/abc", "")

	assertStatus(t, rec, http.StatusBadRequest)
	assertErrorBody(t, rec, "invalid target identifier formatting")
}

func TestDeleteProduct_DBError(t *testing.T) {
	_, mock, mux := newTestAPI(t)

	mock.ExpectExec(`DELETE FROM products`).
		WithArgs(int64(3)).
		WillReturnError(sql.ErrConnDone)

	rec := doRequest(mux, http.MethodDelete, "/api/products/3", "")

	assertStatus(t, rec, http.StatusInternalServerError)
	assertErrorBody(t, rec, "unable to wipe inventory entry database records")
}