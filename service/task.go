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

// TaskFilter describes a page of tasks to list.
type TaskFilter struct {
	Page     int
	PageSize int
	Sort     string
	Order    string
	Name     string
	Status   *int32
	Tag      string
}

func (s *TaskService) GetTasks(ctx context.Context, userID int32, f TaskFilter) ([]*model.Task, int64, error) {
	page, pageSize, sort, order, name := f.Page, f.PageSize, f.Sort, f.Order, f.Name
	q := s.q.Task.WithContext(ctx).Where(s.q.Task.UserID.Eq(userID))

	if f.Status != nil {
		q = q.Where(s.q.Task.Status.Eq(*f.Status))
	}

	if f.Tag != "" {
		tagID, found, err := s.findTagID(ctx, userID, f.Tag)
		if err != nil {
			return nil, 0, err
		}
		if !found {
			return []*model.Task{}, 0, nil
		}
		q = q.Where(s.q.Task.Columns(s.q.Task.ID).In(s.taskIDsWithTag(ctx, tagID)))
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
	if err != nil {
		return nil, 0, err
	}
	if err := s.attachTags(ctx, tasks...); err != nil {
		return nil, 0, err
	}
	return tasks, total, nil
}

// attachTags fills Task.Tags (never nil) for the given tasks with one query.
func (s *TaskService) attachTags(ctx context.Context, tasks ...*model.Task) error {
	if len(tasks) == 0 {
		return nil
	}
	ids := make([]int32, len(tasks))
	for i, t := range tasks {
		ids[i] = t.ID
	}
	byTask, err := loadTags(s.gormDB(ctx), ids)
	if err != nil {
		return err
	}
	for _, t := range tasks {
		t.Tags = byTask[t.ID]
		if t.Tags == nil {
			t.Tags = []string{}
		}
	}
	return nil
}

func (s *TaskService) CreateTask(ctx context.Context, userID int32, task *model.Task) error {
	task.UserID = &userID
	if task.Tags == nil {
		return s.q.Task.WithContext(ctx).Create(task)
	}
	return s.withTx(ctx, func(q *dao.Query, tx *gorm.DB) error {
		if err := q.Task.WithContext(ctx).Create(task); err != nil {
			return err
		}
		return replaceTags(tx, userID, task.ID, task.Tags)
	})
}

func (s *TaskService) UpdateTask(ctx context.Context, userID int32, task *model.Task) error {
	update := func(q *dao.Query) error {
		// Select is required so that zero values (e.g. status 0) are written too.
		info, err := q.Task.WithContext(ctx).
			Select(q.Task.Name, q.Task.Status, q.Task.Description, q.Task.DueDate, q.Task.Priority).
			Where(q.Task.ID.Eq(task.ID), q.Task.UserID.Eq(userID)).
			Updates(task)
		if err != nil {
			return err
		}
		if info.RowsAffected == 0 {
			return ErrTaskNotFound
		}
		return nil
	}

	// nil tags: leave the task's tags unchanged.
	if task.Tags == nil {
		return update(s.q)
	}
	return s.withTx(ctx, func(q *dao.Query, tx *gorm.DB) error {
		if err := update(q); err != nil {
			return err
		}
		return replaceTags(tx, userID, task.ID, task.Tags)
	})
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
	if err != nil {
		return nil, err
	}
	if err := s.attachTags(ctx, task); err != nil {
		return nil, err
	}
	return task, nil
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
