package response

import (
	"encoding/json"
	"fmt"
	"net/http"
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

const (
	jsonContentType = "application/json"
	textContentType = "text/plain"
)

func Error(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", jsonContentType)
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(&ErrorResponse{Error: message})
}

func Ok(w http.ResponseWriter) {
	w.Header().Set("Content-Type", jsonContentType)
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(&OkResponse{Result: "ok"})
}

func Object(w http.ResponseWriter, result interface{}, statusCode int) {
	w.Header().Set("Content-Type", jsonContentType)
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(&ApiResponse{Result: result})
}

func Text(w http.ResponseWriter, str string, statusCode int) {
	w.Header().Set("Content-Type", textContentType)
	w.WriteHeader(statusCode)
	fmt.Fprint(w, str)
}

// Shortcuts for typical response types

// 201
func Created(w http.ResponseWriter, resp interface{}) {
	Object(w, resp, http.StatusCreated)
}

// 204
func Deleted(w http.ResponseWriter) {
	Object(w, nil, http.StatusNoContent)
}

// 404
func NotFound(w http.ResponseWriter, str string) {
	Error(w, str, http.StatusNotFound)
}

// Bad is the shortcut for a response with 400 status code
func Bad(w http.ResponseWriter, str string) {
	Error(w, str, http.StatusBadRequest)
}

// InternalError is the shortcut for a response with 500 status code
func InternalError(w http.ResponseWriter, str string) {
	Error(w, str, http.StatusInternalServerError)
}
