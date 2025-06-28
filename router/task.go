package router

import (
	"github.com/gin-gonic/gin"
	"github/TaskService/handler"
)

func InitTaskRoute(r *gin.Engine) {
	task := r.Group("/")
	task.GET("/tasks", handler.GetTasks)
	task.POST("/tasks", handler.CreateTask)
	task.PUT("/tasks/:id", handler.UpdateTask)
	task.DELETE("/tasks/:id", handler.DeleteTask)
}
