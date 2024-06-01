package controllers

import (
	"errors"
	"fmt"
	"guidance/models"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

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

	// Implement retry logic
	var database *gorm.DB
	var err error
	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		database, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err == nil {
			break
		}
		log.Printf("Failed to connect to database (attempt %d/%d): %v", i+1, maxRetries, err)
		time.Sleep(2 * time.Second) // Wait before retrying
	}
	if err != nil {
		log.Fatalf("Failed to connect to database after %d attempts: %v", maxRetries, err)
	}

	db1 = database

	// Set the maximum number of open connections in the connection pool
	sqlDB, err := db1.DB()
	if err != nil {
		log.Fatalf("Failed to get database connection: %v", err)
	}
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(time.Hour)

	log.Println("Successfully connected to the database")
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

	fmt.Println(rows)

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

	fmt.Println(users)
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
	if result == 2 {
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

	reuser := models.RenrollSchema{Name: input.Name, Phone: input.Phone, Email: input.Email, Date: input.Date, Class: input.Class, Sub: input.Sub, Mentor: "", Renrollment: 0}
	if err := db1.Create(&reuser).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to save data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User data saved successfully"})
}

func DELETE(c *gin.Context) {
	var input deleteSchema
	if err := c.ShouldBindBodyWithJSON(&input); err != nil {
		fmt.Println("2")
		fmt.Println(input)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	fmt.Println("1")
	fmt.Println(input)

	today := time.Now()

	// Subtract 30 days
	thirtyDaysAgo := today.AddDate(0, 0, -30)

	todayFormatted := today.Format("2006-01-02")
	thirtyDaysAgoFormatted := thirtyDaysAgo.Format("2006-01-02")

	// Create a wait group to wait for all Goroutines to finish
	var wg sync.WaitGroup

	// Iterate over the IDs and delete each row concurrently
	for _, id := range input.IDs {
		// Increment the wait group counter
		wg.Add(1)

		// Launch a Goroutine to delete the row
		go func(id uint) {
			defer wg.Done()

			// Construct the delete query
			result := db1.Where("date > ? AND date < ?", thirtyDaysAgoFormatted, todayFormatted).Where("id = ?", id).Delete(&UserSchema{})

			// Check for errors
			if err := result.Error; err != nil {
				// Handle error (e.g., log it)
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
		}(id)
	}

	for _, mentor := range input.Mentors {
		// Skip processing if mentor is empty
		if mentor == "" {
			continue
		}

		wg.Add(1)
		go func(mentor string) {
			defer wg.Done()
			var ment MentorSchema
			if err := db3.Where("name = ?", mentor).First(&ment).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			ment.Onn = ment.Onn - 1
			if err := db1.Save(&ment).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}(mentor)
	}

	// Wait for all Goroutines to finish
	wg.Wait()

	c.JSON(http.StatusOK, gin.H{"message": "User data deleted successfully"})
}
