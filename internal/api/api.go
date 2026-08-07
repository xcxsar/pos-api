package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/xcxsar/pos-api/internal/config"
)

type API struct {
	Cfg *config.ApiConfig
}

func NewAPI(cfg *config.ApiConfig) *API {
	return &API{
		Cfg: cfg,
	}
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	errRes := ErrorResponse{
		Error: msg,
	}

	res, err := json.Marshal(errRes)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Printf("%s", res)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(res)
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	res, err := json.Marshal(payload)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(res)
}
