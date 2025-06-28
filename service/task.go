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

func (s *TaskService) GetTasks(ctx context.Context) ([]*model.Task, error) {
	return s.q.Task.WithContext(ctx).Find()
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
