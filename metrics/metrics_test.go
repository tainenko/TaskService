package metrics

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func newEngine(m *Metrics, token string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(m.Middleware())
	r.GET("/tasks/:id", func(c *gin.Context) { c.Status(http.StatusOK) })
	r.GET("/metrics", m.Handler(token))
	return r
}

func get(r *gin.Engine, path, auth string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, path, nil)
	if auth != "" {
		req.Header.Set("Authorization", auth)
	}
	r.ServeHTTP(w, req)
	return w
}

func TestMiddlewareUsesRouteTemplates(t *testing.T) {
	r := newEngine(New(nil), "")
	get(r, "/tasks/1", "")
	get(r, "/tasks/2", "")
	get(r, "/nope/1234", "")
	get(r, "/nope/5678", "")

	body := get(r, "/metrics", "").Body.String()
	assert.Contains(t, body, `http_requests_total{method="GET",route="/tasks/:id",status="200"} 2`)
	assert.Contains(t, body, `http_requests_total{method="GET",route="unmatched",status="404"} 2`)
	assert.Contains(t, body, `http_request_duration_seconds_bucket{method="GET",route="/tasks/:id"`)
	assert.NotContains(t, body, "/tasks/1", "raw paths must not become labels")
	assert.NotContains(t, body, "/nope/")
	assert.Contains(t, body, "go_goroutines")
}

func TestHandlerToken(t *testing.T) {
	r := newEngine(New(nil), "s3cret")
	assert.Equal(t, http.StatusUnauthorized, get(r, "/metrics", "").Code)
	assert.Equal(t, http.StatusUnauthorized, get(r, "/metrics", "Bearer wrong").Code)
	assert.Equal(t, http.StatusOK, get(r, "/metrics", "Bearer s3cret").Code)
}
