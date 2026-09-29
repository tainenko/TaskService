package service

import (
	"context"
	"errors"
	"github/TaskService/dao"
	"github/TaskService/model"
	"gorm.io/gorm"
	"strings"
)

// ErrTaskNotFound is returned when the target task does not exist.
var ErrTaskNotFound = errors.New("task not found")

type TaskService struct {
	q *dao.Query
}

func NewTaskService(q *dao.Query) *TaskService {
	return &TaskService{q: q}
}

func (s *TaskService) GetTasks(ctx context.Context, userID int32, page, pageSize int, sort, order, name string, status *int32) ([]*model.Task, int64, error) {
	q := s.q.Task.WithContext(ctx).Where(s.q.Task.UserID.Eq(userID))

	if status != nil {
		q = q.Where(s.q.Task.Status.Eq(*status))
	}

	if name != "" {
		q = q.Where(s.q.Task.Name.Like("%" + escapeLike(name) + "%"))
	}

	total, err := q.Count()
	if err != nil {
		return nil, 0, err
	}

	sortedByID := sort == "" || sort == "id"
	if sort != "" {
		if field, ok := s.q.Task.GetFieldByName(sort); ok {
			if order == "asc" {
				q = q.Order(field.Asc())
			} else {
				q = q.Order(field.Desc())
			}
		}
	}
	// Tie-breaker keeps pagination stable when the sort column has duplicates.
	if !sortedByID {
		q = q.Order(s.q.Task.ID.Desc())
	}

	offset := (page - 1) * pageSize
	tasks, err := q.Offset(offset).Limit(pageSize).Find()
	return tasks, total, err
}

func (s *TaskService) CreateTask(ctx context.Context, userID int32, task *model.Task) error {
	task.UserID = &userID
	return s.q.Task.WithContext(ctx).Create(task)
}

func (s *TaskService) UpdateTask(ctx context.Context, userID int32, task *model.Task) error {
	// Select is required so that zero values (e.g. status 0) are written too.
	info, err := s.q.Task.WithContext(ctx).
		Select(s.q.Task.Name, s.q.Task.Status, s.q.Task.Description, s.q.Task.DueDate, s.q.Task.Priority).
		Where(s.q.Task.ID.Eq(task.ID), s.q.Task.UserID.Eq(userID)).
		Updates(task)
	if err != nil {
		return err
	}
	if info.RowsAffected == 0 {
		return ErrTaskNotFound
	}
	return nil
}

func (s *TaskService) DeleteTask(ctx context.Context, userID, id int32) error {
	info, err := s.q.Task.WithContext(ctx).Where(s.q.Task.ID.Eq(id), s.q.Task.UserID.Eq(userID)).Delete()
	if err != nil {
		return err
	}
	if info.RowsAffected == 0 {
		return ErrTaskNotFound
	}
	return nil
}

func (s *TaskService) GetTaskByID(ctx context.Context, userID, id int32) (*model.Task, error) {
	task, err := s.q.Task.WithContext(ctx).Where(s.q.Task.ID.Eq(id), s.q.Task.UserID.Eq(userID)).First()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrTaskNotFound
	}
	return task, err
}

// UpdateTaskStatus changes only the status of a task.
func (s *TaskService) UpdateTaskStatus(ctx context.Context, userID, id, status int32) error {
	info, err := s.q.Task.WithContext(ctx).
		Where(s.q.Task.ID.Eq(id), s.q.Task.UserID.Eq(userID)).
		Update(s.q.Task.Status, status)
	if err != nil {
		return err
	}
	if info.RowsAffected == 0 {
		return ErrTaskNotFound
	}
	return nil
}

var likeEscaper = strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)

// escapeLike escapes LIKE wildcards so user input is matched literally.
func escapeLike(s string) string {
	return likeEscaper.Replace(s)
}
