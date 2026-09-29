package handler

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github/TaskService/service"
)

type fakeTags struct {
	listErr, deleteErr error
}

func (f fakeTags) ListTags(context.Context, int32) ([]service.TagCount, error) {
	return []service.TagCount{{ID: 1, Name: "work", TaskCount: 2}}, f.listErr
}

func (f fakeTags) DeleteTag(context.Context, int32, int32) error { return f.deleteErr }

func TestTagHandler(t *testing.T) {
	tests := []struct {
		name string
		fake fakeTags
		call string
		path string
		want int
	}{
		{"list ok", fakeTags{}, "list", "/tags", http.StatusOK},
		{"list error", fakeTags{listErr: errors.New("boom")}, "list", "/tags", http.StatusInternalServerError},
		{"delete ok", fakeTags{}, "delete", "/tags/1", http.StatusOK},
		{"delete bad id", fakeTags{}, "delete", "/tags/abc", http.StatusBadRequest},
		{"delete not found", fakeTags{deleteErr: service.ErrTagNotFound}, "delete", "/tags/1", http.StatusNotFound},
		{"delete error", fakeTags{deleteErr: errors.New("boom")}, "delete", "/tags/1", http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewTagHandler(tt.fake)
			var code int
			if tt.call == "list" {
				code = doRequest(t, h.ListTags, http.MethodGet, tt.path, "/tags", "").Code
			} else {
				code = doRequest(t, h.DeleteTag, http.MethodDelete, tt.path, "/tags/:id", "").Code
			}
			if code != tt.want {
				t.Errorf("status = %d, want %d", code, tt.want)
			}
		})
	}
}

func TestTaskHandler_InvalidTags(t *testing.T) {
	h := NewTaskHandler(&MockTaskService{})
	w := doRequest(t, h.CreateTask, http.MethodPost, "/tasks", "/tasks", `{"name":"x","status":0,"tags":["ok","  "]}`)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
	w = doRequest(t, h.UpdateTask, http.MethodPut, "/tasks/1", "/tasks/:id", `{"name":"x","status":0,"tags":["ok","  "]}`)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}
