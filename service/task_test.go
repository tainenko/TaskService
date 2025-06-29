package service

import (
	"context"
	"database/sql"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
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
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "task"`).
		WithArgs(sqlmock.AnyArg(), 1).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	s := &TaskService{
		q: q,
	}
	if err := s.DeleteTask(context.Background(), 1); err != nil {
		t.Errorf("DeleteTask() error = %v", err)
	}
}

func TestTaskService_GetTaskByID(t *testing.T) {
	taskID := int32(1)
	expected := &model.Task{
		ID:        taskID,
		Name:      "Test Task",
		Status:    1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	rows := sqlmock.NewRows([]string{"id", "name", "status", "created_at", "updated_at", "deleted_at"}).
		AddRow(expected.ID, expected.Name, expected.Status, expected.CreatedAt, expected.UpdatedAt, nil)

	mock.ExpectQuery(`^SELECT \* FROM "task" WHERE "task"."id" = \$1 AND "task"."deleted_at" IS NULL ORDER BY "task"."id" LIMIT \$2$`).
		WithArgs(1, 1).
		WillReturnRows(rows)

	s := &TaskService{
		q: q,
	}

	actual, err := s.GetTaskByID(context.Background(), taskID)

	assert.NoError(t, err)
	assert.Equal(t, expected.Name, actual.Name)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("GetTaskByID(): %s", err)
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
