package controllers

import (
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

	db4 = database

	// Set the maximum number of open connections in the connection pool
	sqlDB, err := db4.DB()
	if err != nil {
		log.Fatalf("Failed to get database connection: %v", err)
	}
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(time.Hour)

	log.Println("Successfully connected to the database")
}

func RenrollDataGet(c *gin.Context) {
	var wg sync.WaitGroup

	// Query the database to get all users
	rows, err := db4.Model(&RenrollSchema{}).Rows()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	// Collect all rows in a slice
	var userSchemas []RenrollSchema
	for rows.Next() {
		var user RenrollSchema
		if err := db4.ScanRows(rows, &user); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		userSchemas = append(userSchemas, user)
	}

	// Create a channel to receive users from goroutines
	usersChan := make(chan RenrollSchema, len(userSchemas))

	// Create a goroutine for each user schema to process them concurrently
	for _, user := range userSchemas {
		wg.Add(1)
		go func(user RenrollSchema) {
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
	var users []RenrollSchema
	for user := range usersChan {
		users = append(users, user)
	}

	c.JSON(http.StatusOK, users)
}

func RenrollDataPost(c *gin.Context) {
	var input Renroll
	if err := c.ShouldBindBodyWithJSON(&input); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	fmt.Println(input)

	var dte RenrollSchema
	if err := db4.Where("phone=?", input.Phone).First(&dte).Error; err != nil {
		// c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to Find existing user with this Name"})
		// return
	}
	fmt.Println(dte)

	if dte.Date != "" {
		// Parse the string date into a time.Time value
		prevDate, err := time.Parse("2006-01-02", dte.Date) // Assuming the date format is "YYYY-MM-DD"
		if err != nil {
			fmt.Println("done1")
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse re-enrollment date"})
			return
		}

		newDate, err := time.Parse("2006-01-02", input.Date) // Assuming the date format is "YYYY-MM-DD"
		if err != nil {
			fmt.Println("done1")
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse re-enrollment date"})
			return
		}

		duration := newDate.Sub(prevDate)
		// Calculate the difference in days
		diffInDays := int(duration.Hours() / 24)

		if diffInDays < 30 {
			fmt.Println("done")
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "A re-enrollment for this phone number has already been done within the last 30 days."})
			return
		}
	}

	var existingUser RenrollSchema
	if err := db4.Where("name=?", input.Name).First(&existingUser).Error; err != nil {
		// data := models.UserSchema{Name: input.Name, Phone: input.Phone, Email: input.Email, Date: input.Date, Class: input.Class, Sub: input.Sub, Mentor: input.Mentor}
		// if err := db4.Create(&data).Error; err != nil {
		// 	c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to save data"})
		// 	return
		// }

		reuser := models.RenrollSchema{Name: input.Name, Phone: input.Phone, Email: input.Email, Date: input.Date, Class: input.Class, Sub: input.Sub, Mentor: input.Mentor, Renrollment: 1}
		if err := db4.Create(&reuser).Error; err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to save data"})
			return
		}

	} else {
		existingUser.Phone = input.Phone
		existingUser.Email = input.Email
		existingUser.Date = input.Date
		existingUser.Class = input.Class
		existingUser.Sub = input.Sub
		existingUser.Mentor = input.Mentor
		existingUser.Renrollment = existingUser.Renrollment + 1

		if err := db4.Save(&existingUser).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Renrolled successfully"})

}
