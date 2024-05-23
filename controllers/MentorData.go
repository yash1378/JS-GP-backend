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

	db2 = database

	// Set the maximum number of open connections in the connection pool
	sqlDB, err := db2.DB()
	if err != nil {
		log.Fatalf("Failed to get database connection: %v", err)
	}
	sqlDB.SetMaxOpenConns(100)
}

func MentorGet(c *gin.Context) {
	var wg1 sync.WaitGroup

	rows, err := db2.Model(&MentorSchema{}).Rows()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// makes sure to lose all rows data to prevent data leaks
	defer rows.Close()

	var mentorSchemas []MentorSchema
	for rows.Next() {
		var mentor MentorSchema
		// here go routines can't be used as then db would be called by all at the same time
		if err := db2.ScanRows(rows, &mentor); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		mentorSchemas = append(mentorSchemas, mentor)
	}

	mentorsChan := make(chan MentorSchema, len(mentorSchemas))
	for _, mentor := range mentorSchemas {
		wg1.Add(1)
		go func(mentor MentorSchema) {
			defer wg1.Done()
			mentorsChan <- mentor
		}(mentor)
	}

	go func() {
		wg1.Wait()
		close(mentorsChan)
	}()

	var mentors []MentorSchema
	for mentor := range mentorsChan {
		mentors = append(mentors, mentor)
	}

	c.JSON(http.StatusOK, mentors)
}

func MentorPost(c *gin.Context) {
	var input Mentor
	var wg1 sync.WaitGroup

	if err := c.ShouldBindBodyWithJSON(&input); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resultChan := make(chan int, 1)
	wg1.Add(1)

	go func(input Mentor) {
		defer wg1.Done()
		mentor := MentorSchema{}
		if err := db2.Where("name=?", input.Name).First(&mentor).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// No user found, it's safe to proceed with creating a new user
				resultChan <- 2
			} else {
				// Some other error occurred while querying the database
				resultChan <- 3
			}
		} else {
			resultChan <- 2
		}
	}(input)

	go func() {
		wg1.Wait()
		close(resultChan)
	}()
	result := <-resultChan
	if result != 2 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Phone number already exists"})
		return
	} else if result == 3 {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to check existing user"})
		return
	}

	data := models.MentorSchema{Name: input.Name, College: input.College, Date: input.Date}
	if err := db2.Create(&data).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to save data"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Mentor data saved successfully"})

}
