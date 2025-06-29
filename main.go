package main

import (
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"github/TaskService/conf"
	"github/TaskService/dao"
	"github/TaskService/router"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"net/http"
)

func main() {
	env := flag.String("env", "local", "Environment: local|dev|prod")
	flag.Parse()

	config, err := conf.LoadConfig(*env)
	if err != nil {
		panic(err)
	}

	err = setupDB(config.Database)
	if err != nil {
		panic(err)
	}

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

func setupDB(config conf.Database) error {
	dsn := fmt.Sprintf("postgres://%s:%s@%s/%s", config.Username, config.Password, config.Host, config.DBName)
	db, err := gorm.Open(postgres.Open(dsn))
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}
	dao.SetDefault(db)
	return nil
}
