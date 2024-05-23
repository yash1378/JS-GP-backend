package controllers

import (
	"errors"
	"guidance/models"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file")
	}
	host := os.Getenv("DB_HOST")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	port := os.Getenv("DB_PORT")

	dsn := "host=" + host + " user=" + user + " password=" + password + " dbname=" + dbname + " port=" + port
	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	db1 = database

	// Set the maximum number of open connections in the connection pool
	sqlDB, err := db1.DB()
	if err != nil {
		log.Fatalf("Failed to get database connection: %v", err)
	}
	sqlDB.SetMaxOpenConns(100)
}

func UserGet(c *gin.Context) {
	var wg sync.WaitGroup

	// Query the database to get all users
	rows, err := db1.Model(&UserSchema{}).Rows()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	// Collect all rows in a slice
	var userSchemas []UserSchema
	for rows.Next() {
		var user UserSchema
		if err := db1.ScanRows(rows, &user); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		userSchemas = append(userSchemas, user)
	}

	// Create a channel to receive users from goroutines
	usersChan := make(chan UserSchema, len(userSchemas))

	// Create a goroutine for each user schema to process them concurrently
	for _, user := range userSchemas {
		wg.Add(1)
		go func(user UserSchema) {
			defer wg.Done()
			// Simulate processing each user
			usersChan <- user
		}(user)
	}

	// Close the WaitGroup after all goroutines finish
	go func() {
		wg.Wait()
		close(usersChan)
	}()

	// Collect users from the channel and add them to the response slice
	var users []UserSchema
	for user := range usersChan {
		users = append(users, user)
	}

	c.JSON(http.StatusOK, users)
}

func UserPost(c *gin.Context) {
	var input User
	var wg sync.WaitGroup

	if err := c.ShouldBindJSON(&input); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create a channel to receive users from goroutines
	resultChan := make(chan int, 1)
	wg.Add(1)

	go func(input User) {
		defer wg.Done()
		naam := UserSchema{}
		if err := db1.Where("phone = ?", input.Phone).First(&naam).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// No user found, it's safe to proceed with creating a new user
				resultChan <- 1
			} else {
				// Some other error occurred while querying the database
				resultChan <- 3
			}
		} else {
			// User with the same phone number already exists
			resultChan <- 2
		}
	}(input)
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Wait for the result from the goroutine
	result := <-resultChan
	if result != 2 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Phone number already exists"})
		return
	} else if result == 3 {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to check existing user"})
		return
	}
	// If user doesn't exist, proceed with saving the user
	data := models.UserSchema{Name: input.Name, Phone: input.Phone, Email: input.Email, Date: input.Date, Class: input.Class, Sub: input.Sub, Mentor: input.Mentor}
	if err := db1.Create(&data).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to save data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User data saved successfully"})

}
