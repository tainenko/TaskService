package handler

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func GetTasks(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Get all tasks",
	})
}

func CreateTask(c *gin.Context) {
	c.JSON(http.StatusCreated, gin.H{
		"message": "Create new task",
	})
}

func UpdateTask(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{
		"message": "Update task " + id,
	})
}

func DeleteTask(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{
		"message": "Delete task " + id,
	})
}
