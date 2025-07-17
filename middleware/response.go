// Package middleware provides HTTP middleware components for the task service.
package middleware

import error2 "github/TaskService/middleware/error"

// Response represents a standardized API response structure.
// It is used to maintain consistency across all API endpoints.
// T represents the type of data contained in the response.
type Response[T any] struct {
	error2.ErrCode
	// Data contains the actual response payload of type T
	Data T `json:"data"`
}
