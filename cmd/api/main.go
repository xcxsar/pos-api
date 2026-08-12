package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/xcxsar/pos-api/internal/api"
	"github.com/xcxsar/pos-api/internal/config"
	"github.com/xcxsar/pos-api/internal/store/sqlc"
)

func main() {
	if err := godotenv.Load("../../.env"); err != nil {
		log.Fatal(".env file not found")
	}

	port := os.Getenv("PORT")
	dbUrl := os.Getenv("DB_URL")
	jwtSecret := os.Getenv("JWT_SECRET")

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

	apiCfg := &config.ApiConfig{
		Queries:   dbQueries,
		DB:        db,
		JWTSecret: jwtSecret,
	}

	apiApp := api.NewAPI(apiCfg)

	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/users", apiApp.CreateUser)
	mux.HandleFunc("GET /api/users/{userID}", apiApp.GetUserByID)

	mux.HandleFunc("POST /api/login", apiApp.LogIn)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Println("server started, listening at localhost:" + port)
	log.Fatal(server.ListenAndServe())
}
