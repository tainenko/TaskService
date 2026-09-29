package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github/TaskService/response"
)

// CustomRecovery returns a middleware that recovers from any panics, logs the
// panic with its stack trace, and returns a standardized error response.
func CustomRecovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				slog.Error("panic recovered",
					"error", err,
					"method", c.Request.Method,
					"path", c.Request.URL.Path,
					"stack", string(debug.Stack()),
				)
				response.Fail(c, http.StatusInternalServerError, response.ServerError, "Unknown Error Occurred")
			}
		}()

		c.Next() // Process the request
	}
}
