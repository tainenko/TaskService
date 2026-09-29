package service

import (
	"context"

	"github/TaskService/dao"
	"github/TaskService/model"
	"gorm.io/gorm"
)

// CreateTasks creates all tasks atomically: either every task (with its tags)
// is stored or none is. IDs are filled into the given tasks.
func (s *TaskService) CreateTasks(ctx context.Context, userID int32, tasks []*model.Task) error {
	return s.withTx(ctx, func(q *dao.Query, tx *gorm.DB) error {
		for _, task := range tasks {
			task.UserID = &userID
			if err := q.Task.WithContext(ctx).Create(task); err != nil {
				return err
			}
			if task.Tags != nil {
				if err := replaceTags(tx, userID, task.ID, task.Tags); err != nil {
					return err
				}
			}
		}
		return nil
	})
}

// ownedIDs returns the subset of ids that exist and belong to the user.
func ownedIDs(q *dao.Query, ctx context.Context, userID int32, ids []int32) ([]int32, error) {
	tasks, err := q.Task.WithContext(ctx).
		Select(q.Task.ID).
		Where(q.Task.UserID.Eq(userID), q.Task.ID.In(ids...)).
		Find()
	if err != nil {
		return nil, err
	}
	owned := make([]int32, len(tasks))
	for i, t := range tasks {
		owned[i] = t.ID
	}
	return owned, nil
}

// DeleteTasks soft-deletes the user's tasks among ids and returns the IDs that
// were deleted. IDs that do not exist or belong to someone else are skipped.
func (s *TaskService) DeleteTasks(ctx context.Context, userID int32, ids []int32) ([]int32, error) {
	var deleted []int32
	err := s.withTx(ctx, func(q *dao.Query, _ *gorm.DB) error {
		owned, err := ownedIDs(q, ctx, userID, ids)
		if err != nil || len(owned) == 0 {
			return err
		}
		if _, err := q.Task.WithContext(ctx).Where(q.Task.UserID.Eq(userID), q.Task.ID.In(owned...)).Delete(); err != nil {
			return err
		}
		deleted = owned
		return nil
	})
	return deleted, err
}

// UpdateTasksStatus sets the status of the user's tasks among ids and returns
// the IDs that were updated. Unknown or foreign IDs are skipped.
func (s *TaskService) UpdateTasksStatus(ctx context.Context, userID int32, ids []int32, status int32) ([]int32, error) {
	var updated []int32
	err := s.withTx(ctx, func(q *dao.Query, _ *gorm.DB) error {
		owned, err := ownedIDs(q, ctx, userID, ids)
		if err != nil || len(owned) == 0 {
			return err
		}
		if _, err := q.Task.WithContext(ctx).
			Where(q.Task.UserID.Eq(userID), q.Task.ID.In(owned...)).
			Update(q.Task.Status, status); err != nil {
			return err
		}
		updated = owned
		return nil
	})
	return updated, err
}
