package handler

import (
	"bytes"
	"context"
	"github.com/gin-gonic/gin"
	"github/TaskService/model"
	"github/TaskService/service"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

type MockTaskService struct {
}

func (m *MockTaskService) CreateTask(_ context.Context, task *model.Task) error {
	return nil
}

func (m *MockTaskService) GetTasks(_ context.Context, page, pageSize int, sort, order, name string, status *int32) ([]*model.Task, int64, error) {
	return []*model.Task{}, int64(0), nil
}

func (m *MockTaskService) UpdateTask(_ context.Context, task *model.Task) error {
	return nil
}

func (m *MockTaskService) GetTaskByID(_ context.Context, id int32) (*model.Task, error) {
	if id == 404 {
		return nil, service.ErrTaskNotFound
	}
	return &model.Task{ID: id, Name: "t"}, nil
}

func (m *MockTaskService) UpdateTaskStatus(_ context.Context, id, status int32) error {
	if id == 404 {
		return service.ErrTaskNotFound
	}
	return nil
}

func (m *MockTaskService) DeleteTask(_ context.Context, id int32) error {
	return nil
}

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	m.Run()
}

func TestNewTaskHandler(t *testing.T) {
	type args struct {
		taskService TaskServiceInterface
	}
	tests := []struct {
		name string
		args args
		want *TaskHandler
	}{
		{
			name: "Create new task handler",
			args: args{
				taskService: &MockTaskService{},
			},
			want: &TaskHandler{
				taskService: &MockTaskService{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewTaskHandler(tt.args.taskService); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewTaskHandler() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTaskHandler_CreateTask(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	jsonStr := `{"name":"Test Task","status":1}`
	c.Request, _ = http.NewRequest(http.MethodPost, "/tasks", bytes.NewBuffer([]byte(jsonStr)))
	c.Request.Header.Set("Content-Type", "application/json")

	type fields struct {
		taskService TaskServiceInterface
	}
	type args struct {
		c *gin.Context
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		wantStatus int
	}{
		{
			name: "Create task successfully",
			fields: fields{
				taskService: &MockTaskService{},
			},
			args: args{
				c: c,
			},
			wantStatus: http.StatusCreated,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &TaskHandler{
				taskService: tt.fields.taskService,
			}
			h.CreateTask(tt.args.c)
			if w.Code != tt.wantStatus {
				t.Errorf("CreateTask() status = %v, want %v", w.Code, tt.wantStatus)
			}
		})
	}
}

func TestTaskHandler_DeleteTask(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request, _ = http.NewRequest(http.MethodDelete, "/tasks/1", nil)
	c.Params = []gin.Param{{Key: "id", Value: "1"}}

	type fields struct {
		taskService TaskServiceInterface
	}
	type args struct {
		c *gin.Context
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		wantStatus int
	}{
		{
			name: "Delete task successfully",
			fields: fields{
				taskService: &MockTaskService{},
			},
			args: args{
				c: c,
			},
			wantStatus: http.StatusOK,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &TaskHandler{
				taskService: tt.fields.taskService,
			}
			h.DeleteTask(tt.args.c)
			if w.Code != tt.wantStatus {
				t.Errorf("DeleteTask() status = %v, want %v", w.Code, tt.wantStatus)
			}
		})
	}
}

func TestTaskHandler_GetTasks(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request, _ = http.NewRequest(http.MethodGet, "/tasks?page=1&page_size=10", nil)

	type fields struct {
		taskService TaskServiceInterface
	}
	type args struct {
		c *gin.Context
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		wantStatus int
	}{
		{
			name: "Get tasks successfully",
			fields: fields{
				taskService: &MockTaskService{},
			},
			args: args{
				c: c,
			},
			wantStatus: http.StatusOK,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &TaskHandler{
				taskService: tt.fields.taskService,
			}
			h.GetTasks(tt.args.c)
			if w.Code != tt.wantStatus {
				t.Errorf("GetTasks() status = %v, want %v", w.Code, tt.wantStatus)
			}
		})
	}
}

func TestTaskHandler_UpdateTask(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	jsonStr := `{"name":"Updated Task","status":1}`
	c.Request, _ = http.NewRequest(http.MethodPut, "/tasks/1", bytes.NewBuffer([]byte(jsonStr)))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = []gin.Param{{Key: "id", Value: "1"}}

	type fields struct {
		taskService TaskServiceInterface
	}
	type args struct {
		c *gin.Context
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		wantStatus int
	}{
		{
			name: "Update task successfully",
			fields: fields{
				taskService: &MockTaskService{},
			},
			args: args{
				c: c,
			},
			wantStatus: http.StatusOK,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &TaskHandler{
				taskService: tt.fields.taskService,
			}
			h.UpdateTask(tt.args.c)
			if w.Code != tt.wantStatus {
				t.Errorf("UpdateTask() status = %v, %v, want %v", w.Code, w.Body, tt.wantStatus)
			}
		})
	}
}

func doRequest(t *testing.T, h gin.HandlerFunc, method, path, route, body string) *httptest.ResponseRecorder {
	t.Helper()
	r := gin.New()
	r.Handle(method, route, h)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(method, path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	return w
}

func TestTaskHandler_GetTask(t *testing.T) {
	h := NewTaskHandler(&MockTaskService{})
	tests := []struct {
		path string
		want int
	}{
		{"/tasks/1", http.StatusOK},
		{"/tasks/404", http.StatusNotFound},
		{"/tasks/abc", http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			w := doRequest(t, h.GetTask, http.MethodGet, tt.path, "/tasks/:id", "")
			if w.Code != tt.want {
				t.Errorf("status = %d, want %d", w.Code, tt.want)
			}
		})
	}
}

func TestTaskHandler_UpdateTaskStatus(t *testing.T) {
	h := NewTaskHandler(&MockTaskService{})
	tests := []struct {
		name, path, body string
		want             int
	}{
		{"ok", "/tasks/1/status", `{"status":2}`, http.StatusOK},
		{"zero status is valid", "/tasks/1/status", `{"status":0}`, http.StatusOK},
		{"missing status", "/tasks/1/status", `{}`, http.StatusBadRequest},
		{"invalid status", "/tasks/1/status", `{"status":9}`, http.StatusBadRequest},
		{"not found", "/tasks/404/status", `{"status":1}`, http.StatusNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := doRequest(t, h.UpdateTaskStatus, http.MethodPatch, tt.path, "/tasks/:id/status", tt.body)
			if w.Code != tt.want {
				t.Errorf("status = %d, want %d (%s)", w.Code, tt.want, w.Body.String())
			}
		})
	}
}

func TestTaskHandler_UpdateTask_NotFound(t *testing.T) {
	h := NewTaskHandler(&notFoundService{})
	w := doRequest(t, h.UpdateTask, http.MethodPut, "/tasks/1", "/tasks/:id", `{"name":"a","status":0}`)
	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", w.Code)
	}
}

type notFoundService struct{ MockTaskService }

func (notFoundService) UpdateTask(context.Context, *model.Task) error { return service.ErrTaskNotFound }
