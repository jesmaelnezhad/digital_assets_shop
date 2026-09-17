package models

import "time"

// APIResponse is the standard response wrapper for all API endpoints.
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorDetail `json:"error,omitempty"`
}

// PaginatedResponse wraps paginated data.
type PaginatedResponse struct {
	Success    bool        `json:"success"`
	Data       interface{} `json:"data"`
	Total      int         `json:"total"`
	Page       int         `json:"page"`
	PerPage    int         `json:"per_page"`
	TotalPages int         `json:"total_pages"`
}

// ErrorDetail provides structured error information.
type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// NewSuccessResponse creates a success APIResponse.
func NewSuccessResponse(data interface{}) APIResponse {
	return APIResponse{Success: true, Data: data}
}

// NewErrorResponse creates an error APIResponse.
func NewErrorResponse(code, message string, details ...string) APIResponse {
	var d string
	if len(details) > 0 {
		d = details[0]
	}
	return APIResponse{
		Success: false,
		Error:   &ErrorDetail{Code: code, Message: message, Details: d},
	}
}

// NewPaginatedResponse creates a paginated response.
func NewPaginatedResponse(data interface{}, total, page, perPage int) PaginatedResponse {
	totalPages := total / perPage
	if total%perPage > 0 {
		totalPages++
	}
	return PaginatedResponse{
		Success:    true,
		Data:       data,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
	}
}

// HealthResponse for service health checks.
type HealthResponse struct {
	Status    string    `json:"status"`
	Service   string    `json:"service"`
	Version   string    `json:"version"`
	Timestamp time.Time `json:"timestamp"`
}
