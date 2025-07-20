package middleware

import (
	"github.com/gin-gonic/gin"
	"github/TaskService/response"
	"net/http"
)

// CustomRecovery returns a middleware that recovers from any panics and returns a custom response
func CustomRecovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				response.Fail(c, http.StatusInternalServerError, response.ServerError, "Unknown Error Occurred")
			}
		}()

		c.Next() // Process the request
	}
}
