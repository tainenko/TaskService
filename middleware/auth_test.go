package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github/TaskService/auth"
)

func TestAuthRequired(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tm, err := auth.NewTokenManager("0123456789abcdef0123456789abcdef", time.Hour)
	require.NoError(t, err)
	valid, _, _ := tm.Generate(7)

	r := gin.New()
	r.GET("/me", AuthRequired(tm), func(c *gin.Context) {
		id, _ := UserID(c)
		c.JSON(http.StatusOK, gin.H{"id": id})
	})

	tests := []struct {
		name, header string
		want         int
	}{
		{"valid", "Bearer " + valid, http.StatusOK},
		{"lowercase scheme", "bearer " + valid, http.StatusOK},
		{"missing", "", http.StatusUnauthorized},
		{"wrong scheme", "Basic " + valid, http.StatusUnauthorized},
		{"empty token", "Bearer ", http.StatusUnauthorized},
		{"bad token", "Bearer nope", http.StatusUnauthorized},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/me", nil)
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}
			r.ServeHTTP(w, req)
			assert.Equal(t, tt.want, w.Code)
		})
	}
}
