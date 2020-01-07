package api

import (
	"net/http"
	"encoding/json"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

type OkResponse struct {
	Result string `json:"result"`
}

type ApiResponse struct {
	Result interface{} `json:"result"`
}

var jsonContentType = "application/json"

func apiError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", jsonContentType)
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(&ErrorResponse{Error: message})
}

func ok(w http.ResponseWriter) {
	w.Header().Set("Content-Type", jsonContentType)
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(&OkResponse{Result: "ok"})
}

func response(w http.ResponseWriter, result interface{}, statusCode int) {
	w.Header().Set("Content-Type", jsonContentType)
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(&ApiResponse{Result: result})
}