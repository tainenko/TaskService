package service

import (
	"context"
	"github/TaskService/dao/model"
	"github/TaskService/dao/query"
)

type TaskService struct {
	q *query.Query
}

func NewTaskService(q *query.Query) *TaskService {
	return &TaskService{q: q}
}

func (s *TaskService) GetTasks(ctx context.Context, page, pageSize int, sort, order, name string, status *int32) ([]*model.Task, int64, error) {
	q := s.q.Task.WithContext(ctx)

	if status != nil {
		q = q.Where(s.q.Task.Status.Eq(*status))
	}

	if name != "" {
		q = q.Where(s.q.Task.Name.Like("%" + name + "%"))
	}

	total, err := q.Count()
	if err != nil {
		return nil, 0, err
	}

	if sort != "" {
		if field, ok := s.q.Task.GetFieldByName(sort); ok {
			if order == "asc" {
				q = q.Order(field.Asc())
			} else if order == "desc" {
				q = q.Order(field.Desc())
			}
		}
	}

	offset := (page - 1) * pageSize
	tasks, err := q.Offset(offset).Limit(pageSize).Find()
	return tasks, total, err
}

func (s *TaskService) CreateTask(ctx context.Context, task *model.Task) error {
	return s.q.Task.WithContext(ctx).Create(task)
}

func (s *TaskService) UpdateTask(ctx context.Context, task *model.Task) error {
	_, err := s.q.Task.WithContext(ctx).Where(s.q.Task.ID.Eq(task.ID)).Updates(task)
	return err
}

func (s *TaskService) DeleteTask(ctx context.Context, id int32) error {
	_, err := s.q.Task.WithContext(ctx).Where(s.q.Task.ID.Eq(id)).Delete()
	return err
}

func (s *TaskService) GetTaskByID(ctx context.Context, id int32) (*model.Task, error) {
	return s.q.Task.WithContext(ctx).Where(s.q.Task.ID.Eq(id)).First()
}
