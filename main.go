package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"github/TaskService/dao"
	"github/TaskService/router"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"net/http"
)

func main() {
	readConfig()

	setupDB()

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

func setupDB() {
	username := viper.GetString("Database.Username")
	password := viper.GetString("Database.Password")
	host := viper.GetString("Database.Host")
	dbName := viper.GetString("Database.DBName")
	dsn := fmt.Sprintf("postgres://%s:%s@%s/%s", username, password, host, dbName)

	db, err := gorm.Open(postgres.Open(dsn))
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to database: %v", err))
	}
	dao.SetDefault(db)
}

func readConfig() {
	vp := viper.New()
	vp.SetConfigName("config")
	vp.AddConfigPath("conf/")
	vp.SetConfigType("yaml")
	err := vp.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
	fmt.Printf("Using config:%+v\n", vp.AllSettings())
}
