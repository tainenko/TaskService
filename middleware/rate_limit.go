package middleware

import (
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github/TaskService/ratelimit"
	"github/TaskService/response"
)

// RateLimitByIP rejects requests with 429 once a client IP exceeds its budget.
// The IP comes from gin's ClientIP, so configure trusted proxies on the engine
// or the limit can be bypassed with a spoofed X-Forwarded-For header.
func RateLimitByIP(l *ratelimit.KeyedLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		if ok, wait := l.Allow(c.ClientIP()); !ok {
			TooManyRequests(c, wait)
			return
		}
		c.Next()
	}
}

// TooManyRequests aborts with 429 and a Retry-After header.
func TooManyRequests(c *gin.Context, wait time.Duration) {
	c.Header("Retry-After", strconv.Itoa(int(math.Ceil(wait.Seconds()))))
	response.Fail(c, http.StatusTooManyRequests, response.TooManyRequests, "too many requests, retry later")
}
