package router

import (
	"github.com/gin-gonic/gin"
	"github/TaskService/auth"
	"github/TaskService/dao"
	"github/TaskService/handler"
	"github/TaskService/middleware"
	"github/TaskService/ratelimit"
	"github/TaskService/service"
	"gorm.io/gorm"
)

// SetAuthRoute registers the public registration and login endpoints, limited
// per client IP by ipLimiter. failures throttles repeated failed logins per account.
func SetAuthRoute(r *gin.Engine, db *gorm.DB, tokens *auth.TokenManager, ipLimiter, failures *ratelimit.KeyedLimiter) {
	h := handler.NewAuthHandler(service.NewUserService(db), tokens, failures)
	g := r.Group("/auth", middleware.RateLimitByIP(ipLimiter))
	g.POST("/register", h.Register)
	g.POST("/login", h.Login)
}

// SetTaskRoute registers the task endpoints, which require a valid access token.
func SetTaskRoute(r *gin.Engine, tokens *auth.TokenManager) {
	s := service.NewTaskService(dao.Q)
	taskHandler := handler.NewTaskHandler(s)
	task := r.Group("/", middleware.AuthRequired(tokens))
	// Static batch routes take precedence over the /tasks/:id patterns.
	task.POST("/tasks/batch", taskHandler.CreateTasks)
	task.DELETE("/tasks/batch", taskHandler.DeleteTasks)
	task.PATCH("/tasks/batch/status", taskHandler.UpdateTasksStatus)
	task.GET("/tasks", taskHandler.GetTasks)
	task.GET("/tasks/:id", taskHandler.GetTask)
	task.POST("/tasks", taskHandler.CreateTask)
	task.PUT("/tasks/:id", taskHandler.UpdateTask)
	task.PATCH("/tasks/:id/status", taskHandler.UpdateTaskStatus)
	task.DELETE("/tasks/:id", taskHandler.DeleteTask)

	tagHandler := handler.NewTagHandler(s)
	task.GET("/tags", tagHandler.ListTags)
	task.DELETE("/tags/:id", tagHandler.DeleteTag)
}
