// Package httputil provides standard JSON response helpers for the TelePromptr
// API. It defines consistent envelope formats for success responses (with
// optional pagination) and error responses with structured error details.
package httputil

import (
	"encoding/json"
	"net/http"
)

// ContentTypeJSON is the Content-Type header value for JSON responses.
const ContentTypeJSON = "application/json"

// Pagination holds pagination metadata for list responses.
type Pagination struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// SuccessResponse is the standard envelope for successful API responses.
type SuccessResponse struct {
	Data       any         `json:"data"`
	Pagination *Pagination `json:"pagination,omitempty"`
}

// ErrorDetail holds structured information about an API error.
type ErrorDetail struct {
	Code    string   `json:"code"`
	Message string   `json:"message"`
	Details []string `json:"details,omitempty"`
}

// ErrorResponse is the standard envelope for error API responses.
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// JSON writes any value as a JSON response with the given status code.
func JSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", ContentTypeJSON)
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		// At this point headers are already sent; best effort logging only.
		http.Error(w, "", http.StatusInternalServerError)
	}
}

// Success writes a success response with the given data and HTTP 200 OK.
func Success(w http.ResponseWriter, data any) {
	JSON(w, http.StatusOK, SuccessResponse{Data: data})
}

// SuccessWithStatus writes a success response with the given data and
// custom HTTP status code.
func SuccessWithStatus(w http.ResponseWriter, status int, data any) {
	JSON(w, status, SuccessResponse{Data: data})
}

// SuccessWithPagination writes a paginated success response with HTTP 200 OK.
// It computes TotalPages from total and perPage automatically.
func SuccessWithPagination(w http.ResponseWriter, data any, page, perPage, total int) {
	totalPages := 0
	if perPage > 0 {
		totalPages = (total + perPage - 1) / perPage
	}

	JSON(w, http.StatusOK, SuccessResponse{
		Data: data,
		Pagination: &Pagination{
			Page:       page,
			PerPage:    perPage,
			Total:      total,
			TotalPages: totalPages,
		},
	})
}

// Err writes an error response with the given HTTP status code, error code,
// and human-readable message.
func Err(w http.ResponseWriter, status int, code, message string) {
	JSON(w, status, ErrorResponse{
		Error: ErrorDetail{
			Code:    code,
			Message: message,
		},
	})
}

// ErrWithDetails writes an error response that includes additional detail
// strings (e.g., field-level validation errors).
func ErrWithDetails(w http.ResponseWriter, status int, code, message string, details []string) {
	JSON(w, status, ErrorResponse{
		Error: ErrorDetail{
			Code:    code,
			Message: message,
			Details: details,
		},
	})
}

// NotFound writes a 404 Not Found error response.
func NotFound(w http.ResponseWriter, message string) {
	Err(w, http.StatusNotFound, "not_found", message)
}

// BadRequest writes a 400 Bad Request error response.
func BadRequest(w http.ResponseWriter, message string) {
	Err(w, http.StatusBadRequest, "bad_request", message)
}

// InternalError writes a 500 Internal Server Error response.
func InternalError(w http.ResponseWriter, message string) {
	Err(w, http.StatusInternalServerError, "internal_error", message)
}

// Unauthorized writes a 401 Unauthorized error response.
func Unauthorized(w http.ResponseWriter, message string) {
	Err(w, http.StatusUnauthorized, "unauthorized", message)
}

// Forbidden writes a 403 Forbidden error response.
func Forbidden(w http.ResponseWriter, message string) {
	Err(w, http.StatusForbidden, "forbidden", message)
}
