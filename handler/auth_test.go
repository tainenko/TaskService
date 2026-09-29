package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github/TaskService/model"
	"github/TaskService/service"
)

type fakeUsers struct {
	registerErr, authErr error
}

func (f fakeUsers) Register(_ context.Context, email, _ string) (*model.User, error) {
	return &model.User{ID: 1, Email: email}, f.registerErr
}

func (f fakeUsers) Authenticate(_ context.Context, email, _ string) (*model.User, error) {
	return &model.User{ID: 1, Email: email}, f.authErr
}

type fakeTokens struct{ err error }

func (f fakeTokens) Generate(int32) (string, time.Time, error) {
	return "tok", time.Now().Add(time.Hour), f.err
}

func TestAuthHandler(t *testing.T) {
	good := `{"email":"a@example.com","password":"password123"}`
	tests := []struct {
		name   string
		users  fakeUsers
		tokens fakeTokens
		call   func(*AuthHandler) gin.HandlerFunc
		body   string
		want   int
	}{
		{"register ok", fakeUsers{}, fakeTokens{}, func(h *AuthHandler) gin.HandlerFunc { return h.Register }, good, http.StatusCreated},
		{"register bad email", fakeUsers{}, fakeTokens{}, func(h *AuthHandler) gin.HandlerFunc { return h.Register }, `{"email":"nope","password":"password123"}`, http.StatusBadRequest},
		{"register short password", fakeUsers{}, fakeTokens{}, func(h *AuthHandler) gin.HandlerFunc { return h.Register }, `{"email":"a@example.com","password":"short"}`, http.StatusBadRequest},
		{"register too long password", fakeUsers{}, fakeTokens{}, func(h *AuthHandler) gin.HandlerFunc { return h.Register }, `{"email":"a@example.com","password":"` + `xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"}`, http.StatusBadRequest},
		{"register duplicate", fakeUsers{registerErr: service.ErrEmailTaken}, fakeTokens{}, func(h *AuthHandler) gin.HandlerFunc { return h.Register }, good, http.StatusConflict},
		{"register db error", fakeUsers{registerErr: errors.New("boom")}, fakeTokens{}, func(h *AuthHandler) gin.HandlerFunc { return h.Register }, good, http.StatusInternalServerError},
		{"login ok", fakeUsers{}, fakeTokens{}, func(h *AuthHandler) gin.HandlerFunc { return h.Login }, good, http.StatusOK},
		{"login wrong password", fakeUsers{authErr: service.ErrInvalidCredentials}, fakeTokens{}, func(h *AuthHandler) gin.HandlerFunc { return h.Login }, good, http.StatusUnauthorized},
		{"login token failure", fakeUsers{}, fakeTokens{err: errors.New("boom")}, func(h *AuthHandler) gin.HandlerFunc { return h.Login }, good, http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewAuthHandler(tt.users, tt.tokens, nil)
			w := doRequest(t, tt.call(h), http.MethodPost, "/x", "/x", tt.body)
			if w.Code != tt.want {
				t.Errorf("status = %d, want %d (%s)", w.Code, tt.want, w.Body.String())
			}
		})
	}
}

func TestTaskHandler_RequiresUser(t *testing.T) {
	h := NewTaskHandler(&MockTaskService{})
	w := httptestNoUser(h.GetTasks)
	if w != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w)
	}
}

func httptestNoUser(h gin.HandlerFunc) int {
	r := gin.New()
	r.GET("/tasks", h)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/tasks", nil)
	r.ServeHTTP(w, req)
	return w.Code
}

type stubLimiter struct {
	allowed       bool
	fails, resets int
}

func (s *stubLimiter) Allowed(string) (bool, time.Duration) { return s.allowed, 30 * time.Second }
func (s *stubLimiter) Fail(string)                          { s.fails++ }
func (s *stubLimiter) Reset(string)                         { s.resets++ }

func TestAuthHandler_LoginFailureLimiter(t *testing.T) {
	body := `{"email":"a@example.com","password":"password123"}`

	t.Run("locked account is rejected before authenticating", func(t *testing.T) {
		lim := &stubLimiter{allowed: false}
		h := NewAuthHandler(fakeUsers{}, fakeTokens{}, lim)
		w := doRequest(t, h.Login, http.MethodPost, "/x", "/x", body)
		if w.Code != http.StatusTooManyRequests {
			t.Fatalf("status = %d, want 429", w.Code)
		}
		if w.Header().Get("Retry-After") != "30" {
			t.Errorf("Retry-After = %q, want 30", w.Header().Get("Retry-After"))
		}
		if lim.fails != 0 || lim.resets != 0 {
			t.Errorf("limiter state must not change on a blocked attempt")
		}
	})

	t.Run("bad password counts as a failure", func(t *testing.T) {
		lim := &stubLimiter{allowed: true}
		h := NewAuthHandler(fakeUsers{authErr: service.ErrInvalidCredentials}, fakeTokens{}, lim)
		w := doRequest(t, h.Login, http.MethodPost, "/x", "/x", body)
		if w.Code != http.StatusUnauthorized || lim.fails != 1 || lim.resets != 0 {
			t.Errorf("status=%d fails=%d resets=%d", w.Code, lim.fails, lim.resets)
		}
	})

	t.Run("server error is not counted as a failure", func(t *testing.T) {
		lim := &stubLimiter{allowed: true}
		h := NewAuthHandler(fakeUsers{authErr: errors.New("boom")}, fakeTokens{}, lim)
		doRequest(t, h.Login, http.MethodPost, "/x", "/x", body)
		if lim.fails != 0 {
			t.Errorf("fails = %d, want 0", lim.fails)
		}
	})

	t.Run("success resets the counter", func(t *testing.T) {
		lim := &stubLimiter{allowed: true}
		h := NewAuthHandler(fakeUsers{}, fakeTokens{}, lim)
		w := doRequest(t, h.Login, http.MethodPost, "/x", "/x", body)
		if w.Code != http.StatusOK || lim.resets != 1 {
			t.Errorf("status=%d resets=%d", w.Code, lim.resets)
		}
	})
}
