package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github/TaskService/middleware"
	"github/TaskService/model"
	"github/TaskService/response"
	"github/TaskService/service"
)

// MaxBatchSize is the largest number of items accepted by a batch endpoint.
const MaxBatchSize = 100

type BatchCreateRequest struct {
	Tasks []TaskRequest `json:"tasks" binding:"required,min=1,max=100,dive"`
}

type BatchIDsRequest struct {
	IDs []int32 `json:"ids" binding:"required,min=1,max=100,dive,gt=0"`
}

type BatchStatusRequest struct {
	IDs    []int32 `json:"ids" binding:"required,min=1,max=100,dive,gt=0"`
	Status *int32  `json:"status" binding:"required,oneof=0 1 2"`
}

// BatchResult reports which IDs were processed and which were not found
// (missing or owned by someone else).
type BatchResult struct {
	Processed []int32 `json:"processed"`
	NotFound  []int32 `json:"not_found"`
}

func newBatchResult(requested, processed []int32) BatchResult {
	done := make(map[int32]bool, len(processed))
	for _, id := range processed {
		done[id] = true
	}
	res := BatchResult{Processed: []int32{}, NotFound: []int32{}}
	for _, id := range requested {
		if done[id] {
			res.Processed = append(res.Processed, id)
		} else {
			res.NotFound = append(res.NotFound, id)
		}
	}
	return res
}

// dedupe removes duplicate IDs, preserving order.
func dedupe(ids []int32) []int32 {
	seen := make(map[int32]bool, len(ids))
	out := make([]int32, 0, len(ids))
	for _, id := range ids {
		if !seen[id] {
			seen[id] = true
			out = append(out, id)
		}
	}
	return out
}

// CreateTasks creates up to MaxBatchSize tasks atomically.
func (h *TaskHandler) CreateTasks(c *gin.Context) {
	userID, ok := middleware.UserID(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, response.Unauthorized, "unauthorized")
		return
	}

	var req BatchCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.InvalidPayload, err.Error())
		return
	}

	tasks := make([]*model.Task, len(req.Tasks))
	for i, item := range req.Tasks {
		tags, err := service.NormalizeTags(item.Tags)
		if err != nil {
			response.Fail(c, http.StatusBadRequest, response.InvalidPayload, fmt.Sprintf("tasks[%d]: %s", i, err))
			return
		}
		tasks[i] = &model.Task{
			Name: item.Name, Status: item.Status, Description: item.Description,
			DueDate: item.DueDate, Priority: item.Priority, Tags: tags,
		}
	}

	if err := h.taskService.CreateTasks(c.Request.Context(), userID, tasks); err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CreateTaskErr, "failed to create tasks")
		return
	}

	ids := make([]int32, len(tasks))
	for i, t := range tasks {
		ids[i] = t.ID
	}
	response.Created(c, gin.H{"ids": ids})
}

// DeleteTasks deletes the given tasks; unknown IDs are reported, not fatal.
func (h *TaskHandler) DeleteTasks(c *gin.Context) {
	userID, ok := middleware.UserID(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, response.Unauthorized, "unauthorized")
		return
	}

	var req BatchIDsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.InvalidPayload, err.Error())
		return
	}
	ids := dedupe(req.IDs)

	deleted, err := h.taskService.DeleteTasks(c.Request.Context(), userID, ids)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, response.DeleteTaskErr, "failed to delete tasks")
		return
	}
	response.Success(c, newBatchResult(ids, deleted))
}

// UpdateTasksStatus sets the status of the given tasks; unknown IDs are reported, not fatal.
func (h *TaskHandler) UpdateTasksStatus(c *gin.Context) {
	userID, ok := middleware.UserID(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, response.Unauthorized, "unauthorized")
		return
	}

	var req BatchStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.InvalidPayload, err.Error())
		return
	}
	ids := dedupe(req.IDs)

	updated, err := h.taskService.UpdateTasksStatus(c.Request.Context(), userID, ids, *req.Status)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, response.UpdateTaskErr, "failed to update tasks")
		return
	}
	response.Success(c, newBatchResult(ids, updated))
}
