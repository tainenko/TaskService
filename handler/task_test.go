package handler

import (
	"github.com/gin-gonic/gin"
	"reflect"
	"testing"
)

func TestNewTaskHandler(t *testing.T) {
	type args struct {
		taskService TaskServiceInterface
	}
	tests := []struct {
		name string
		args args
		want *TaskHandler
	}{
		// TODO: Add test cases.
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
	type fields struct {
		taskService TaskServiceInterface
	}
	type args struct {
		c *gin.Context
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &TaskHandler{
				taskService: tt.fields.taskService,
			}
			h.CreateTask(tt.args.c)
		})
	}
}

func TestTaskHandler_DeleteTask(t *testing.T) {
	type fields struct {
		taskService TaskServiceInterface
	}
	type args struct {
		c *gin.Context
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &TaskHandler{
				taskService: tt.fields.taskService,
			}
			h.DeleteTask(tt.args.c)
		})
	}
}

func TestTaskHandler_GetTasks(t *testing.T) {
	type fields struct {
		taskService TaskServiceInterface
	}
	type args struct {
		c *gin.Context
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &TaskHandler{
				taskService: tt.fields.taskService,
			}
			h.GetTasks(tt.args.c)
		})
	}
}

func TestTaskHandler_UpdateTask(t *testing.T) {
	type fields struct {
		taskService TaskServiceInterface
	}
	type args struct {
		c *gin.Context
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &TaskHandler{
				taskService: tt.fields.taskService,
			}
			h.UpdateTask(tt.args.c)
		})
	}
}
