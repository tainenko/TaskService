package router

import (
	"github.com/gin-gonic/gin"
	"github/TaskService/auth"
	"github/TaskService/dao"
	"github/TaskService/handler"
	"github/TaskService/middleware"
	"github/TaskService/service"
	"gorm.io/gorm"
)

// SetAuthRoute registers the public registration and login endpoints.
func SetAuthRoute(r *gin.Engine, db *gorm.DB, tokens *auth.TokenManager) {
	h := handler.NewAuthHandler(service.NewUserService(db), tokens)
	g := r.Group("/auth")
	g.POST("/register", h.Register)
	g.POST("/login", h.Login)
}

// SetTaskRoute registers the task endpoints, which require a valid access token.
func SetTaskRoute(r *gin.Engine, tokens *auth.TokenManager) {
	s := service.NewTaskService(dao.Q)
	taskHandler := handler.NewTaskHandler(s)
	task := r.Group("/", middleware.AuthRequired(tokens))
	task.GET("/tasks", taskHandler.GetTasks)
	task.GET("/tasks/:id", taskHandler.GetTask)
	task.POST("/tasks", taskHandler.CreateTask)
	task.PUT("/tasks/:id", taskHandler.UpdateTask)
	task.PATCH("/tasks/:id/status", taskHandler.UpdateTaskStatus)
	task.DELETE("/tasks/:id", taskHandler.DeleteTask)
}
