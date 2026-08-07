package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/xcxsar/pos-api/internal/store/sqlc"
)

type apiConfig struct {
	queries   *sqlc.Queries
	db        *sql.DB
	jwtSecret string
}

func main() {
	if err := godotenv.Load("../../.env"); err != nil {
		log.Fatal(".env file not found")
	}

	port := os.Getenv("PORT")
	dbUrl := os.Getenv("DB_URL")

	db, err := sql.Open("postgres", dbUrl)
	if err != nil {
		log.Fatal("failed to open DB pool:", err)
	}

	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatal("failed to connect to the database:", err)
	}

	log.Println("connection to database established")

	dbQueries := sqlc.New(db)

	var apiCfg apiConfig

	apiCfg.queries = dbQueries
	apiCfg.db = db

	mux := http.NewServeMux()

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Println("server started, listening at localhost:" + port)
	log.Fatal(server.ListenAndServe())
}
