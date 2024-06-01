package controllers

import (
	"errors"
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

	db5 = database

	// Set the maximum number of open connections in the connection pool
	sqlDB, err := db5.DB()
	if err != nil {
		log.Fatalf("Failed to get database connection: %v", err)
	}
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(time.Hour)

	log.Println("Successfully connected to the database")
}

func OwnerGet(c *gin.Context) {
	var wg sync.WaitGroup

	// Query the database to get all owners
	rows, err := db5.Model(&OwnerSchema{}).Rows()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	// Collect all rows in a slice
	var ownerSchemas []OwnerSchema
	for rows.Next() {
		var owner OwnerSchema
		if err := db5.ScanRows(rows, &owner); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		ownerSchemas = append(ownerSchemas, owner)
	}

	// Create a channel to receive owners from goroutines
	ownersChan := make(chan OwnerSchema, len(ownerSchemas))

	// Create a goroutine for each owner schema to process them concurrently
	for _, owner := range ownerSchemas {
		wg.Add(1)
		go func(owner OwnerSchema) {
			defer wg.Done()
			// Simulate processing each owner
			ownersChan <- owner
		}(owner)
	}

	// Close the WaitGroup after all goroutines finish
	go func() {
		wg.Wait()
		close(ownersChan)
	}()

	// Collect owners from the channel and add them to the response slice
	var owners []OwnerSchema
	for owner := range ownersChan {
		owners = append(owners, owner)
	}

	c.JSON(http.StatusOK, owners)
}

func OwnerPost(c *gin.Context) {
	var input Owner
	var wg sync.WaitGroup

	if err := c.ShouldBindJSON(&input); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create a channel to receive owners from goroutines
	resultChan := make(chan int, 1)
	wg.Add(1)

	go func(input Owner) {
		defer wg.Done()
		naam := OwnerSchema{}
		if err := db5.Where("email = ?", input.Email).First(&naam).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// No owner found, it's safe to proceed with creating a new owner
				resultChan <- 1
			} else {
				// Some other error occurred while querying the database
				resultChan <- 3
			}
		} else {
			// owner with the same phone number already exists
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
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to check existing owner"})
		return
	}

	hashed, err := hashPassword(input.Password)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to save data"})
		return
	}
	// If owner doesn't exist, proceed with saving the owner
	data := models.OwnerSchema{Email: input.Email, Password: hashed, OwnerName: input.OwnerName}
	if err := db5.Create(&data).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to save data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "owner data saved successfully"})

}
