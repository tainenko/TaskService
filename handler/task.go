package handler

import (
	"context"
	"github.com/gin-gonic/gin"
	"github/TaskService/model"
	"net/http"
	"strconv"
)

type TaskServiceInterface interface {
	GetTasks(ctx context.Context, page, pageSize int, sort, order, name string, status *int32) ([]*model.Task, int64, error)
	CreateTask(ctx context.Context, task *model.Task) error
	UpdateTask(ctx context.Context, task *model.Task) error
	DeleteTask(ctx context.Context, id int32) error
}

type TaskHandler struct {
	taskService TaskServiceInterface
}

func NewTaskHandler(taskService TaskServiceInterface) *TaskHandler {
	return &TaskHandler{taskService: taskService}
}

type TaskRequest struct {
	Name   string `json:"name" binding:"required"`
	Status int32  `json:"status" binding:"oneof=0 1"`
}

type PaginationResponse struct {
	TotalRecords int64 `json:"total_records"`
	CurrentPage  int   `json:"current_page"`
	TotalPages   int   `json:"total_pages"`
	NextPage     *int  `json:"next_page"`
	PrevPage     *int  `json:"prev_page"`
}

type TaskListRequest struct {
	Page     int    `form:"page,default=1"`
	PageSize int    `form:"pageSize,default=10"`
	Sort     string `form:"sort,default=id"`
	Order    string `form:"order,default=desc"`
	Name     string `form:"name"`
	Status   *int32 `form:"status"`
}

func (h *TaskHandler) GetTasks(c *gin.Context) {
	var req TaskListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tasks, total, err := h.taskService.GetTasks(c.Request.Context(), req.Page, req.PageSize, req.Sort, req.Order, req.Name, req.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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

	c.JSON(http.StatusOK, gin.H{
		"data":       tasks,
		"pagination": pagination,
	})
}

func (h *TaskHandler) CreateTask(c *gin.Context) {
	var req TaskRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task := &model.Task{
		Name:   req.Name,
		Status: req.Status,
	}

	if err := h.taskService.CreateTask(c.Request.Context(), task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, task.ID)
}

func (h *TaskHandler) UpdateTask(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task id"})
		return
	}

	var req TaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task := &model.Task{
		ID:     int32(id),
		Name:   req.Name,
		Status: req.Status,
	}

	if err := h.taskService.UpdateTask(c.Request.Context(), task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task updated successfully"})
}

func (h *TaskHandler) DeleteTask(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task id"})
		return
	}

	if err := h.taskService.DeleteTask(c.Request.Context(), int32(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task deleted successfully"})
}
