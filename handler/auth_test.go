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
			h := NewAuthHandler(tt.users, tt.tokens)
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
