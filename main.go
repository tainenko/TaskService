package main

import (
	"github.com/gin-gonic/gin"
	"github/TaskService/router"
	"net/http"
)

func main() {
	// Create default gin router
	r := gin.Default()

	// Define a basic route
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Welcome to the API",
		})
	})
	router.InitTaskRoute(r)

	// Run the server on port 8080
	r.Run(":8080")
}
