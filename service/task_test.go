package service

import (
	"context"
	"github/TaskService/dao/model"
	"github/TaskService/dao/query"
	"reflect"
	"testing"
)

func TestNewTaskService(t *testing.T) {
	type args struct {
		q *query.Query
	}
	tests := []struct {
		name string
		args args
		want *TaskService
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewTaskService(tt.args.q); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewTaskService() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTaskService_CreateTask(t *testing.T) {
	type fields struct {
		q *query.Query
	}
	type args struct {
		ctx  context.Context
		task *model.Task
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &TaskService{
				q: tt.fields.q,
			}
			if err := s.CreateTask(tt.args.ctx, tt.args.task); (err != nil) != tt.wantErr {
				t.Errorf("CreateTask() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTaskService_DeleteTask(t *testing.T) {
	type fields struct {
		q *query.Query
	}
	type args struct {
		ctx context.Context
		id  int32
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &TaskService{
				q: tt.fields.q,
			}
			if err := s.DeleteTask(tt.args.ctx, tt.args.id); (err != nil) != tt.wantErr {
				t.Errorf("DeleteTask() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTaskService_GetTaskByID(t *testing.T) {
	type fields struct {
		q *query.Query
	}
	type args struct {
		ctx context.Context
		id  int32
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.Task
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &TaskService{
				q: tt.fields.q,
			}
			got, err := s.GetTaskByID(tt.args.ctx, tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetTaskByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetTaskByID() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTaskService_GetTasks(t *testing.T) {
	type fields struct {
		q *query.Query
	}
	type args struct {
		ctx      context.Context
		page     int
		pageSize int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*model.Task
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &TaskService{
				q: tt.fields.q,
			}
			got, err := s.GetTasks(tt.args.ctx, tt.args.page, tt.args.pageSize)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetTasks() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetTasks() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTaskService_UpdateTask(t *testing.T) {
	type fields struct {
		q *query.Query
	}
	type args struct {
		ctx  context.Context
		task *model.Task
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &TaskService{
				q: tt.fields.q,
			}
			if err := s.UpdateTask(tt.args.ctx, tt.args.task); (err != nil) != tt.wantErr {
				t.Errorf("UpdateTask() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
