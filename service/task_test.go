package service

import (
	"context"
	"database/sql"
	"github.com/DATA-DOG/go-sqlmock"
	"github/TaskService/dao/model"
	"github/TaskService/dao/query"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"os"
	"reflect"
	"testing"
	"time"
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

var (
	mockDB *sql.DB
	mock   sqlmock.Sqlmock
	q      *query.Query
)

func TestMain(m *testing.M) {
	var err error
	mockDB, mock, err = sqlmock.New()
	if err != nil {
		log.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mockDB.Close()

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: mockDB,
	}), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to gorm DB: %s", err)
	}

	q = query.Use(gormDB)
	os.Exit(m.Run())
}

func TestTaskService_CreateTask(t *testing.T) {
	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "task"`).
		WithArgs("name", 1, nil).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).AddRow(1, time.Now(), time.Now()))
	mock.ExpectCommit()

	s := &TaskService{
		q: q,
	}
	if err := s.CreateTask(context.Background(), &model.Task{Name: "name", Status: 1}); err != nil {
		t.Errorf("CreateTask() error = %v", err)
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
