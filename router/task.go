package router

import (
	"github.com/gin-gonic/gin"
	"github/TaskService/dao"
	"github/TaskService/handler"
	"github/TaskService/service"
)

func SetTaskRoute(r *gin.Engine) {
	s := service.NewTaskService(dao.Q)
	taskHandler := handler.NewTaskHandler(s)
	task := r.Group("/")
	task.GET("/tasks", taskHandler.GetTasks)
	task.POST("/tasks", taskHandler.CreateTask)
	task.PUT("/tasks/:id", taskHandler.UpdateTask)
	task.DELETE("/tasks/:id", taskHandler.DeleteTask)
}
