package config

import (
	"database/sql"

	"github.com/xcxsar/pos-api/internal/store/sqlc"
)

type ApiConfig struct {
	Queries   *sqlc.Queries
	DB        *sql.DB
	JWTSecret string
}
