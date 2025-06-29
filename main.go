package main

import (
	"github.com/gin-gonic/gin"
	"github/TaskService/dao"
	"github/TaskService/router"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"net/http"
)

func main() {
	gormdb, _ := gorm.Open(postgres.Open("postgres://postgres:admin@localhost:5432/postgres"))
	dao.SetDefault(gormdb)

	// Create default gin router
	r := gin.Default()

	// Define a basic route
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Welcome to the API",
		})
	})
	router.SetTaskRoute(r)

	// Run the server on port 8080
	r.Run(":8080")
}
