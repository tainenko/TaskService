package router

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSetTaskRoute(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a new Gin engine
	r := gin.Default()
	SetTaskRoute(r)

	// Test GET /tasks
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/tasks", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("GET /tasks returned %v, expected %v", w.Code, http.StatusOK)
	}

	// Test POST /tasks
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/tasks", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Errorf("POST /tasks returned %v, expected %v", w.Code, http.StatusCreated)
	}

	// Test PUT /tasks/1
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("PUT", "/tasks/1", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("PUT /tasks/1 returned %v, expected %v", w.Code, http.StatusOK)
	}

	// Test DELETE /tasks/1
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("DELETE", "/tasks/1", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("DELETE /tasks/1 returned %v, expected %v", w.Code, http.StatusOK)
	}
}
