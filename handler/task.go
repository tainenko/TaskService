package handler

import (
	"context"
	"errors"
	"github.com/creasty/defaults"
	"github.com/gin-gonic/gin"
	"github/TaskService/middleware"
	"github/TaskService/model"
	"github/TaskService/response"
	"github/TaskService/service"
	"net/http"
	"strconv"
	"time"
)

type TaskServiceInterface interface {
	GetTaskByID(ctx context.Context, userID, id int32) (*model.Task, error)
	UpdateTaskStatus(ctx context.Context, userID, id, status int32) error
	GetTasks(ctx context.Context, userID int32, filter service.TaskFilter) ([]*model.Task, int64, error)
	CreateTask(ctx context.Context, userID int32, task *model.Task) error
	UpdateTask(ctx context.Context, userID int32, task *model.Task) error
	DeleteTask(ctx context.Context, userID, id int32) error
}

type TaskHandler struct {
	taskService TaskServiceInterface
}

func NewTaskHandler(taskService TaskServiceInterface) *TaskHandler {
	return &TaskHandler{taskService: taskService}
}

// Task status values: 0 = todo, 1 = done, 2 = in progress.
// Priority values: 0 = low, 1 = medium, 2 = high.
type TaskRequest struct {
	Name        string     `json:"name" binding:"required,max=255"`
	Status      int32      `json:"status" binding:"oneof=0 1 2"`
	Description string     `json:"description"`
	DueDate     *time.Time `json:"due_date"`
	Priority    int32      `json:"priority" binding:"oneof=0 1 2"`
	// Tags replaces the task's tags. Omit it to leave them unchanged; [] clears them.
	Tags []string `json:"tags"`
}

type TaskStatusRequest struct {
	Status *int32 `json:"status" binding:"required,oneof=0 1 2"`
}

type PaginationResponse struct {
	TotalRecords int64 `json:"total_records"`
	CurrentPage  int   `json:"current_page"`
	TotalPages   int   `json:"total_pages"`
	NextPage     *int  `json:"next_page"`
	PrevPage     *int  `json:"prev_page"`
}

type TaskListRequest struct {
	Page     int    `form:"page" default:"1" binding:"gte=1"`
	PageSize int    `form:"pageSize" default:"10" binding:"gte=10,lte=100"`
	Sort     string `form:"sort" default:"id"`
	Order    string `form:"order" default:"desc"`
	Name     string `form:"name"`
	Status   *int32 `form:"status" binding:"omitempty,oneof=0 1 2"`
	Tag      string `form:"tag" binding:"max=50"`
}

type TaskListResponse struct {
	Page  PaginationResponse `json:"page"`
	Tasks []*model.Task      `json:"tasks"`
}

func (h *TaskHandler) GetTasks(c *gin.Context) {
	userID, ok := middleware.UserID(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, response.Unauthorized, "unauthorized")
		return
	}

	req := TaskListRequest{}

	if err := defaults.Set(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.InvalidParam, err.Error())
		return
	}

	if err := c.ShouldBindQuery(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.InvalidParam, err.Error())
		return
	}

	tasks, total, err := h.taskService.GetTasks(c.Request.Context(), userID, service.TaskFilter{
		Page: req.Page, PageSize: req.PageSize, Sort: req.Sort, Order: req.Order,
		Name: req.Name, Status: req.Status, Tag: req.Tag,
	})
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, response.GetTasksErr, err.Error())
		return
	}

	totalPages := int((total + int64(req.PageSize) - 1) / int64(req.PageSize))
	var nextPage *int
	var prevPage *int

	if req.Page < totalPages {
		next := req.Page + 1
		nextPage = &next
	}
	if req.Page > 1 {
		prev := req.Page - 1
		prevPage = &prev
	}

	pagination := PaginationResponse{
		TotalRecords: total,
		CurrentPage:  req.Page,
		TotalPages:   totalPages,
		NextPage:     nextPage,
		PrevPage:     prevPage,
	}

	response.Success(c, TaskListResponse{Tasks: tasks, Page: pagination})
}

type TaskResponse struct {
	ID int32 `json:"id"`
}

func (h *TaskHandler) CreateTask(c *gin.Context) {
	userID, ok := middleware.UserID(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, response.Unauthorized, "unauthorized")
		return
	}

	var req TaskRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.InvalidPayload, err.Error())
		return
	}

	tags, err := service.NormalizeTags(req.Tags)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, response.InvalidPayload, err.Error())
		return
	}

	task := &model.Task{
		Name:        req.Name,
		Status:      req.Status,
		Description: req.Description,
		DueDate:     req.DueDate,
		Priority:    req.Priority,
		Tags:        tags,
	}

	if err := h.taskService.CreateTask(c.Request.Context(), userID, task); err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CreateTaskErr, err.Error())
		return
	}

	response.Created(c, TaskResponse{ID: task.ID})
}

func (h *TaskHandler) UpdateTask(c *gin.Context) {
	userID, ok := middleware.UserID(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, response.Unauthorized, "unauthorized")
		return
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil || id <= 0 {
		response.Fail(c, http.StatusBadRequest, response.InvalidParam, "invalid task id")
		return
	}

	var req TaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.InvalidPayload, err.Error())
		return
	}

	tags, err := service.NormalizeTags(req.Tags)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, response.InvalidPayload, err.Error())
		return
	}

	task := &model.Task{
		ID:          int32(id),
		Name:        req.Name,
		Status:      req.Status,
		Description: req.Description,
		DueDate:     req.DueDate,
		Priority:    req.Priority,
		Tags:        tags,
	}

	if err := h.taskService.UpdateTask(c.Request.Context(), userID, task); err != nil {
		if errors.Is(err, service.ErrTaskNotFound) {
			response.Fail(c, http.StatusNotFound, response.TaskNotFound, err.Error())
			return
		}
		response.Fail(c, http.StatusInternalServerError, response.UpdateTaskErr, err.Error())
		return
	}

	response.Success(c, "Task updated successfully")
}

func (h *TaskHandler) DeleteTask(c *gin.Context) {
	userID, ok := middleware.UserID(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, response.Unauthorized, "unauthorized")
		return
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil || id <= 0 {
		response.Fail(c, http.StatusBadRequest, response.InvalidParam, "invalid task id")
		return
	}

	if err := h.taskService.DeleteTask(c.Request.Context(), userID, int32(id)); err != nil {
		if errors.Is(err, service.ErrTaskNotFound) {
			response.Fail(c, http.StatusNotFound, response.TaskNotFound, err.Error())
			return
		}
		response.Fail(c, http.StatusInternalServerError, response.DeleteTaskErr, err.Error())
		return
	}

	response.Success(c, "Task deleted successfully")
}

func (h *TaskHandler) GetTask(c *gin.Context) {
	userID, ok := middleware.UserID(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, response.Unauthorized, "unauthorized")
		return
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil || id <= 0 {
		response.Fail(c, http.StatusBadRequest, response.InvalidParam, "invalid task id")
		return
	}

	task, err := h.taskService.GetTaskByID(c.Request.Context(), userID, int32(id))
	if err != nil {
		if errors.Is(err, service.ErrTaskNotFound) {
			response.Fail(c, http.StatusNotFound, response.TaskNotFound, err.Error())
			return
		}
		response.Fail(c, http.StatusInternalServerError, response.GetTaskErr, err.Error())
		return
	}

	response.Success(c, task)
}

func (h *TaskHandler) UpdateTaskStatus(c *gin.Context) {
	userID, ok := middleware.UserID(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, response.Unauthorized, "unauthorized")
		return
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil || id <= 0 {
		response.Fail(c, http.StatusBadRequest, response.InvalidParam, "invalid task id")
		return
	}

	var req TaskStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.InvalidPayload, err.Error())
		return
	}

	if err := h.taskService.UpdateTaskStatus(c.Request.Context(), userID, int32(id), *req.Status); err != nil {
		if errors.Is(err, service.ErrTaskNotFound) {
			response.Fail(c, http.StatusNotFound, response.TaskNotFound, err.Error())
			return
		}
		response.Fail(c, http.StatusInternalServerError, response.UpdateTaskErr, err.Error())
		return
	}

	response.Success(c, "Task status updated successfully")
}
