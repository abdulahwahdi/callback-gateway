// Package response gives every HTTP handler in this service one consistent
// JSON envelope, so dashboard/frontend clients only ever parse one shape.
package response

import "github.com/labstack/echo/v4"

// Envelope is the standard response body for every endpoint in this service.
type Envelope struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
	Meta    any    `json:"meta,omitempty"`
}

// Pagination describes pagination metadata attached to list endpoints.
type Pagination struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalData  int64 `json:"total_data"`
	TotalPages int   `json:"total_pages"`
}

// OK writes a successful envelope.
func OK(c echo.Context, code int, message string, data any) error {
	return c.JSON(code, Envelope{Success: true, Message: message, Data: data})
}

// OKPaginated writes a successful envelope with pagination metadata.
func OKPaginated(c echo.Context, code int, message string, data any, page, limit int, total int64) error {
	totalPages := 0
	if limit > 0 {
		totalPages = int((total + int64(limit) - 1) / int64(limit))
	}
	return c.JSON(code, Envelope{
		Success: true,
		Message: message,
		Data:    data,
		Meta: Pagination{
			Page:       page,
			Limit:      limit,
			TotalData:  total,
			TotalPages: totalPages,
		},
	})
}

// Err writes a failure envelope.
func Err(c echo.Context, code int, message string) error {
	return c.JSON(code, Envelope{Success: false, Message: message})
}
