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
	expectedTasks := []*model.Task{
		{
			ID:        1,
			Name:      "Task 1",
			Status:    1,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        2,
			Name:      "Task 2",
			Status:    0,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	rows := sqlmock.NewRows([]string{"id", "name", "status", "created_at", "updated_at", "deleted_at"})
	for _, task := range expectedTasks {
		rows.AddRow(task.ID, task.Name, task.Status, task.CreatedAt, task.UpdatedAt, nil)
	}

	mock.ExpectQuery(`^SELECT count\(\*\) FROM "task" WHERE "task"."deleted_at" IS NULL$`).
		WithArgs().
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	mock.ExpectQuery(`^SELECT \* FROM "task" WHERE "task"."deleted_at" IS NULL ORDER BY "task"."id" DESC LIMIT \$1$`).
		WithArgs(10).
		WillReturnRows(rows)

	s := &TaskService{
		q: q,
	}

	tasks, total, err := s.GetTasks(context.Background(), 1, 10, "id", "desc", "", nil)

	assert.NoError(t, err)
	assert.Equal(t, len(expectedTasks), len(tasks))
	assert.Equal(t, total, int64(2))
	for i, task := range tasks {
		assert.Equal(t, expectedTasks[i].Name, task.Name)
		assert.Equal(t, expectedTasks[i].Status, task.Status)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("GetTasks(): %s", err)
	}
}

func TestTaskService_UpdateTask(t *testing.T) {
	taskID := int32(1)
	updated := &model.Task{
		ID:        taskID,
		Name:      "Test Task",
		Status:    1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mock.ExpectBegin()
	mock.ExpectExec(`^UPDATE "task" SET "name"=\$1,"status"=\$2,"created_at"=\$3,"updated_at"=\$4 WHERE "task"."id" = \$5 AND "task"."deleted_at" IS NULL AND "id" = \$6$`).
		WithArgs(
			updated.Name,
			updated.Status,
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			updated.ID,
			updated.ID,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	s := &TaskService{
		q: q,
	}

	err := s.UpdateTask(context.Background(), updated)

	assert.NoError(t, err)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("UpdateTask(): %s", err)
	}
}
