package main

import (
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load("../../.env"); err != nil {
		log.Fatal(".env file not found")
	}

	port := os.Getenv("PORT")

	mux := http.NewServeMux()

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Println("Server started, listening at localhost:" + port)
	log.Fatal(server.ListenAndServe())
}
