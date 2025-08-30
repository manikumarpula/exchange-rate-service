package models

import (
	"encoding/json"
	"net/http"
	"time"
)

// APIResponse represents a standard API response wrapper
type APIResponse struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Error     string      `json:"error,omitempty"`
	Message   string      `json:"message,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
	Meta      *Meta       `json:"meta,omitempty"`
}

// Meta represents metadata for pagination or additional info
type Meta struct {
	Page       int `json:"page,omitempty"`
	PerPage    int `json:"per_page,omitempty"`
	Total      int `json:"total,omitempty"`
	TotalPages int `json:"total_pages,omitempty"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Success   bool      `json:"success"`
	Error     string    `json:"error"`
	Code      string    `json:"code,omitempty"`
	Timestamp time.Time `json:"timestamp"`
	Details   string    `json:"details,omitempty"`
}

// SuccessResponse creates a successful API response
func SuccessResponse(data interface{}, message string) *APIResponse {
	return &APIResponse{
		Success:   true,
		Data:      data,
		Message:   message,
		Timestamp: time.Now(),
	}
}

// ErrorResponse creates an error API response
func NewErrorResponse(err string, code string, details string) *ErrorResponse {
	return &ErrorResponse{
		Success:   false,
		Error:     err,
		Code:      code,
		Details:   details,
		Timestamp: time.Now(),
	}
}

// WriteJSON writes a JSON response to the HTTP response writer
func WriteJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}
}

// WriteSuccess writes a successful JSON response
func WriteSuccess(w http.ResponseWriter, data interface{}, message string) {
	response := SuccessResponse(data, message)
	WriteJSON(w, http.StatusOK, response)
}

// WriteError writes an error JSON response
func WriteError(w http.ResponseWriter, statusCode int, err string, code string, details string) {
	response := NewErrorResponse(err, code, details)
	WriteJSON(w, statusCode, response)
}

// WriteBadRequest writes a bad request error response
func WriteBadRequest(w http.ResponseWriter, message string) {
	WriteError(w, http.StatusBadRequest, "Bad Request", "BAD_REQUEST", message)
}

// WriteNotFound writes a not found error response
func WriteNotFound(w http.ResponseWriter, message string) {
	WriteError(w, http.StatusNotFound, "Not Found", "NOT_FOUND", message)
}

// WriteInternalError writes an internal server error response
func WriteInternalError(w http.ResponseWriter, message string) {
	WriteError(w, http.StatusInternalServerError, "Internal Server Error", "INTERNAL_ERROR", message)
}
