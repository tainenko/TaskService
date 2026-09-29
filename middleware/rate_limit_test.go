package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github/TaskService/ratelimit"
	"time"
)

func TestRateLimitByIP(t *testing.T) {
	gin.SetMode(gin.TestMode)
	l := ratelimit.NewKeyedLimiter(ratelimit.PerMinute(1), 2, time.Hour)
	r := gin.New()
	// Trust no proxy: the client IP must be the socket address, not X-Forwarded-For.
	_ = r.SetTrustedProxies(nil)
	r.GET("/", RateLimitByIP(l), func(c *gin.Context) { c.Status(http.StatusOK) })

	call := func(remote, xff string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = remote
		if xff != "" {
			req.Header.Set("X-Forwarded-For", xff)
		}
		r.ServeHTTP(w, req)
		return w
	}

	assert.Equal(t, http.StatusOK, call("1.1.1.1:1000", "").Code)
	assert.Equal(t, http.StatusOK, call("1.1.1.1:1001", "").Code)
	w := call("1.1.1.1:1002", "")
	assert.Equal(t, http.StatusTooManyRequests, w.Code)
	assert.NotEmpty(t, w.Header().Get("Retry-After"))

	// Spoofing X-Forwarded-For must not grant a fresh budget.
	assert.Equal(t, http.StatusTooManyRequests, call("1.1.1.1:1003", "9.9.9.9").Code)

	// Another client is unaffected.
	assert.Equal(t, http.StatusOK, call("2.2.2.2:1000", "").Code)
}
