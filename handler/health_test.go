package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type fakePinger struct{ err error }

func (f fakePinger) PingContext(context.Context) error { return f.err }

func TestHealthHandler(t *testing.T) {
	tests := []struct {
		name       string
		pingErr    error
		call       func(*HealthHandler, *gin.Context)
		wantStatus int
	}{
		{"healthz ignores db state", errors.New("down"), (*HealthHandler).Healthz, http.StatusOK},
		{"readyz ok", nil, (*HealthHandler).Readyz, http.StatusOK},
		{"readyz db down", errors.New("down"), (*HealthHandler).Readyz, http.StatusServiceUnavailable},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request, _ = http.NewRequest(http.MethodGet, "/", nil)
			tt.call(NewHealthHandler(fakePinger{tt.pingErr}), c)
			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}
