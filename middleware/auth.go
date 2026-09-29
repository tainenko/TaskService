package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github/TaskService/auth"
	"github/TaskService/response"
)

// UserIDKey is the gin context key holding the authenticated user's ID (int32).
const UserIDKey = "userID"

// AuthRequired validates the Bearer token and stores the user ID in the context.
func AuthRequired(tokens *auth.TokenManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		scheme, token, ok := strings.Cut(c.GetHeader("Authorization"), " ")
		if !ok || !strings.EqualFold(scheme, "Bearer") || token == "" {
			response.Fail(c, http.StatusUnauthorized, response.Unauthorized, "missing or malformed authorization header")
			return
		}

		userID, err := tokens.Parse(strings.TrimSpace(token))
		if err != nil {
			response.Fail(c, http.StatusUnauthorized, response.Unauthorized, "invalid or expired token")
			return
		}

		c.Set(UserIDKey, userID)
		c.Next()
	}
}

// UserID returns the authenticated user's ID set by AuthRequired.
func UserID(c *gin.Context) (int32, bool) {
	id, ok := c.Get(UserIDKey)
	if !ok {
		return 0, false
	}
	userID, ok := id.(int32)
	return userID, ok
}
