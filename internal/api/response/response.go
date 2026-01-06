package response

import (
	"encoding/json"
	"net/http"
)

// ErrorResponse represents an API error response.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
	Code    int    `json:"code"`
}

// SuccessResponse represents a successful API response with data.
type SuccessResponse struct {
	Data interface{} `json:"data"`
}

// PaginatedResponse represents a paginated API response.
type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalCount int         `json:"total_count"`
	TotalPages int         `json:"total_pages"`
}

// JSON writes a JSON response with the given status code.
func JSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		}
	}
}

// Success writes a successful JSON response.
func Success(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusOK, SuccessResponse{Data: data})
}

// Created writes a 201 Created response.
func Created(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusCreated, SuccessResponse{Data: data})
}

// NoContent writes a 204 No Content response.
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// Error writes an error response with the given status code.
func Error(w http.ResponseWriter, status int, err error) {
	JSON(w, status, ErrorResponse{
		Error:   http.StatusText(status),
		Message: err.Error(),
		Code:    status,
	})
}

// BadRequest writes a 400 Bad Request response.
func BadRequest(w http.ResponseWriter, err error) {
	Error(w, http.StatusBadRequest, err)
}

// NotFound writes a 404 Not Found response.
func NotFound(w http.ResponseWriter, err error) {
	Error(w, http.StatusNotFound, err)
}

// InternalError writes a 500 Internal Server Error response.
func InternalError(w http.ResponseWriter, err error) {
	Error(w, http.StatusInternalServerError, err)
}

// ServiceUnavailable writes a 503 Service Unavailable response.
func ServiceUnavailable(w http.ResponseWriter, err error) {
	Error(w, http.StatusServiceUnavailable, err)
}

// Paginated writes a paginated response.
func Paginated(w http.ResponseWriter, data interface{}, page, pageSize, totalCount int) {
	totalPages := (totalCount + pageSize - 1) / pageSize
	if totalPages < 1 {
		totalPages = 1
	}

	JSON(w, http.StatusOK, PaginatedResponse{
		Data:       data,
		Page:       page,
		PageSize:   pageSize,
		TotalCount: totalCount,
		TotalPages: totalPages,
	})
}
