package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github/TaskService/dao"
	"gorm.io/gen"
	"gorm.io/gen/field"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	MaxTagsPerTask = 10
	MaxTagLength   = 50
)

var (
	ErrInvalidTag  = errors.New("invalid tag")
	ErrTagNotFound = errors.New("tag not found")
)

// TagCount is a tag with the number of (non-deleted) tasks using it.
type TagCount struct {
	ID        int32  `json:"id"`
	Name      string `json:"name"`
	TaskCount int64  `json:"task_count"`
}

// NormalizeTags trims, lowercases and de-duplicates tag names, preserving order.
// nil stays nil (meaning "unchanged"); an empty non-nil slice means "no tags".
func NormalizeTags(in []string) ([]string, error) {
	if in == nil {
		return nil, nil
	}
	out := make([]string, 0, len(in))
	seen := make(map[string]bool, len(in))
	for _, raw := range in {
		name := strings.ToLower(strings.TrimSpace(raw))
		if name == "" {
			return nil, fmt.Errorf("%w: tag must not be empty", ErrInvalidTag)
		}
		if len([]rune(name)) > MaxTagLength {
			return nil, fmt.Errorf("%w: tag longer than %d characters", ErrInvalidTag, MaxTagLength)
		}
		if !seen[name] {
			seen[name] = true
			out = append(out, name)
		}
	}
	if len(out) > MaxTagsPerTask {
		return nil, fmt.Errorf("%w: at most %d tags per task", ErrInvalidTag, MaxTagsPerTask)
	}
	return out, nil
}

// gormDB returns a plain *gorm.DB. NewDB drops the Task model (and its
// soft-delete scope) that UnderlyingDB carries, which would otherwise be
// applied to queries on other tables.
func (s *TaskService) gormDB(ctx context.Context) *gorm.DB {
	return s.q.Task.WithContext(ctx).UnderlyingDB().Session(&gorm.Session{NewDB: true})
}

// loadTags returns the tag names for each task ID.
func loadTags(db *gorm.DB, taskIDs []int32) (map[int32][]string, error) {
	var rows []struct {
		TaskID int32
		Name   string
	}
	err := db.Table("task_tag").
		Select("task_tag.task_id AS task_id, tag.name AS name").
		Joins("JOIN tag ON tag.id = task_tag.tag_id").
		Where("task_tag.task_id IN ?", taskIDs).
		Order("tag.name").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make(map[int32][]string, len(taskIDs))
	for _, r := range rows {
		out[r.TaskID] = append(out[r.TaskID], r.Name)
	}
	return out, nil
}

// replaceTags makes tags the exact tag set of the task, creating missing tags.
// It must run inside the caller's transaction.
func replaceTags(tx *gorm.DB, userID, taskID int32, tags []string) error {
	if err := tx.Exec("DELETE FROM task_tag WHERE task_id = ?", taskID).Error; err != nil {
		return err
	}
	if len(tags) == 0 {
		return nil
	}

	type tagRow struct {
		ID     int32  `gorm:"column:id;primaryKey"`
		UserID int32  `gorm:"column:user_id"`
		Name   string `gorm:"column:name"`
	}
	toCreate := make([]tagRow, len(tags))
	for i, name := range tags {
		toCreate[i] = tagRow{UserID: userID, Name: name}
	}
	if err := tx.Table("tag").Clauses(clause.OnConflict{DoNothing: true}).Create(&toCreate).Error; err != nil {
		return err
	}

	var ids []int32
	if err := tx.Table("tag").Where("user_id = ? AND name IN ?", userID, tags).Pluck("id", &ids).Error; err != nil {
		return err
	}
	links := make([]map[string]any, len(ids))
	for i, id := range ids {
		links[i] = map[string]any{"task_id": taskID, "tag_id": id}
	}
	return tx.Table("task_tag").Create(&links).Error
}

// withTx runs fn against a transactional Query and DB.
func (s *TaskService) withTx(ctx context.Context, fn func(q *dao.Query, tx *gorm.DB) error) error {
	return s.gormDB(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(dao.Use(tx), tx)
	})
}

// findTagID looks up the user's tag by (normalized) name.
func (s *TaskService) findTagID(ctx context.Context, userID int32, name string) (int32, bool, error) {
	var ids []int32
	err := s.gormDB(ctx).Table("tag").
		Where("user_id = ? AND name = ?", userID, strings.ToLower(strings.TrimSpace(name))).
		Limit(1).Pluck("id", &ids).Error
	if err != nil || len(ids) == 0 {
		return 0, false, err
	}
	return ids[0], true, nil
}

// taskIDsWithTag is the sub-query SELECT task_id FROM task_tag WHERE tag_id = ?.
func (s *TaskService) taskIDsWithTag(ctx context.Context, tagID int32) gen.SubQuery {
	var tt gen.DO
	tt.UseDB(s.gormDB(ctx))
	tt.UseTable("task_tag")
	// Unscoped: task_tag has no deleted_at column to filter on.
	return tt.Unscoped().Select(field.NewInt32("task_tag", "task_id")).
		Where(field.NewInt32("task_tag", "tag_id").Eq(tagID))
}

// ListTags returns the user's tags with usage counts, ordered by name.
func (s *TaskService) ListTags(ctx context.Context, userID int32) ([]TagCount, error) {
	tags := []TagCount{}
	err := s.gormDB(ctx).Table("tag").
		Select("tag.id AS id, tag.name AS name, COUNT(task.id) AS task_count").
		Joins("LEFT JOIN task_tag ON task_tag.tag_id = tag.id").
		Joins("LEFT JOIN task ON task.id = task_tag.task_id AND task.deleted_at IS NULL").
		Where("tag.user_id = ?", userID).
		Group("tag.id").
		Order("tag.name").
		Scan(&tags).Error
	return tags, err
}

// DeleteTag removes a tag and detaches it from all tasks.
func (s *TaskService) DeleteTag(ctx context.Context, userID, id int32) error {
	res := s.gormDB(ctx).Exec("DELETE FROM tag WHERE id = ? AND user_id = ?", id, userID)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrTagNotFound
	}
	return nil
}
