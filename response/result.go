package response

import "github.com/gin-gonic/gin"

// Result represents the standard API response structure
type Result struct {
	Code    ErrCode     `json:"code"` // Error code for the response
	Message string      `json:"msg"`  // Message describing the result
	Status  int         `json:"-"`    // HTTP status code (not exposed in JSON)
	Data    interface{} `json:"data"` // Payload data of the response
}

// Error implements the error interface and returns the error message
func (r *Result) Error() string {
	return r.Message
}

// New creates a new Result instance with the specified code, message, and data
func New(code ErrCode, message string, data interface{}) *Result {
	return &Result{
		Code:    code,
		Message: message,
		Status:  200,
		Data:    data,
	}
}

// WithError updates the Result with error information and returns the modified Result
func (r *Result) WithError(status int, code ErrCode, message string) *Result {
	r.Status = status
	r.Code = code
	r.Message = message
	return r
}

// WithData sets the data field of the Result and returns the modified Result
func (r *Result) WithData(data interface{}) *Result {
	r.Data = data
	return r
}

// Fail sends a JSON error response with the specified status, code, and message
func Fail(c *gin.Context, status int, code ErrCode, message string) {
	c.JSON(status, Result{
		Code:    code,
		Message: message,
		Data:    nil,
	})
	c.Abort()
}

// Success sends a JSON success response with the specified data
func Success(c *gin.Context, data interface{}) {
	c.JSON(200, Result{
		Code:    0,
		Message: "",
		Data:    data,
	})
}
