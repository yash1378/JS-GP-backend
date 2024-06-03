package controllers

import (
	"errors"
	"fmt"
	"guidance/models"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
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

	db2 = database

	// Set the maximum number of open connections in the connection pool
	sqlDB, err := db2.DB()
	if err != nil {
		log.Fatalf("Failed to get database connection: %v", err)
	}
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(time.Hour)

	log.Println("Successfully connected to the database")
}

func hashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

func comparePassword(hashedPassword, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		return err
	}
	return nil
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

// Remove the repititive saving of data
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
		if err := db2.Where("phone=?", input.Phone).First(&mentor).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// No user found, it's safe to proceed with creating a new user
				resultChan <- 1
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
	if result == 2 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Phone number already exists"})
		return
	} else if result == 3 {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to check existing user"})
		return
	}

	data := models.MentorSchema{Name: input.Name, College: input.College, Date: input.Date, Phone: input.Phone}
	if err := db2.Create(&data).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to save data"})
		return
	}

	// this is to store creds
	wg1.Add(1)
	newChan := make(chan int, 1)

	go func(input Mentor) {
		defer wg1.Done()
		cred := MentorLogin{}
		if err := db2.Where("email=?", input.Email).First(&cred).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// No user found, it's safe to proceed with creating a new user
				newChan <- 1
			} else {
				// Some other error occurred while querying the database
				newChan <- 3
			}
		} else {
			fmt.Println(cred)
			newChan <- 2
		}
	}(input)

	go func() {
		wg1.Wait()
		close(newChan)
	}()

	result = <-newChan
	if result == 2 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Email already exists2"})
		return
	} else if result == 3 {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to check existing user"})
		return
	}

	hashed, err := hashPassword(input.Password)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to save data"})
		return
	}

	otherdata := models.MentorLogin{Email: input.Email, Password: hashed, MentorName: input.Name}
	if err := db2.Create(&otherdata).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to save data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Mentor data saved successfully"})
}

// updating mentors capacity
func MentorStudentUpdate(c *gin.Context) {
	var input UpdateCount
	fmt.Println(c)
	if err := c.ShouldBindBodyWithJSON(&input); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		fmt.Println(input)
		return
	}

	fmt.Println(input)

	cred := MentorSchema{}
	if err := db2.Where("name=?", input.MentorName).First(&cred).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to Find existing user with this Name"})
		return
	}

	// number, err := strconv.Atoi(input.StudentCount)
	number := input.StudentCount

	// if err != nil {
	// 	c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
	// 	return
	// }

	w := cred.Onn //getting the current value of student
	if number < w {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to save data"})
		return
	}
	cred.Handle = number //increasing it by the value of students came in request

	if err := db2.Save(&cred).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to save data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Mentor data Updated successfully"})
}

func MentorInfoUpdate(c *gin.Context) {
	phone := c.Param("phone")

	var input MentorUpdateInfo
	if err := c.ShouldBindBodyWithJSON(&input); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user MentorSchema
	var cred MentorLogin
	if err := db2.Where("phone=?", phone).First(&user).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to Find existing user with this PhoneNo"})
		return
	}

	naam := user.Name

	if err := db2.Where("mentor_name=?", naam).First(&cred).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to Find existing user with this Name"})
		return
	}
	// log.Fatalln(cred)

	user.Name = input.Name
	user.Phone = input.PhoneNo
	cred.MentorName = input.Name
	cred.Email = input.Email

	// log.Fatalln("executed")

	if err := db2.Save(&user).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to save data"})
		return
	}
	if err := db2.Save(&cred).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to save data"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Mentor data Updated successfully"})
}

func ChangePassword(c *gin.Context) {
	var input PasswordChange

	if err := c.ShouldBindBodyWithJSON(&input); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// fmt.Println(input.Email)

	var user MentorLogin
	if err := db2.Where("LOWER(email) = ?", strings.ToLower(strings.TrimSpace(input.Email))).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "User with this email not found"})
		} else {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to find existing user with this email"})
		}
		return
	}

	err := comparePassword(user.Password, input.OldPassword)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	} else {
		hashedNew, err := hashPassword(input.NewPassword)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to save data"})
			return
		}
		user.Password = hashedNew
		if err := db2.Save(&user).Error; err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to save data"})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{"message": "Mentor Password updated successfully"})

}

func FinalMentor(c *gin.Context) {
	var input FinalMentorSchema
	fmt.Println(1)
	if err := c.ShouldBindBodyWithJSON(&input); err != nil {
		fmt.Println(input)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	fmt.Println(2)
	selectedStudentCount := len(input.IDs)
	var ment MentorSchema
	if err := db2.Where("name = ?", input.MentorName).First(&ment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	fmt.Println(3)

	if ment.Onn+selectedStudentCount > ment.Handle {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "mentor can't handle this much"})
		return
	}
	fmt.Println(4)
	ment.Onn = ment.Onn + selectedStudentCount
	ment.Total = ment.Total + selectedStudentCount

	if err := db2.Save(&ment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	fmt.Println(5)
	// Collect all rows in a slice
	var mu sync.Mutex

	fmt.Println(input)

	// Initialize students slice
	var students []UserSchema
	var phoneNumbers []string
	var wg sync.WaitGroup
	// Iterate over the IDs and fetch records concurrently
	fmt.Println(6)
	for _, id := range input.IDs {
		// Increment the wait group counter
		wg.Add(1)

		// Launch a Goroutine to fetch the record
		go func(id int) {
			defer wg.Done()

			// Fetch the record with the given ID
			var user UserSchema
			result := db2.First(&user, id)

			// Check for errors
			if err := result.Error; err != nil {
				// Handle error (e.g., log it)
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			// Lock the mutex before appending to the students slice
			mu.Lock()
			defer mu.Unlock()

			// Append the fetched record to the students slice
			students = append(students, user)
			fmt.Println(user.Phone)
			phoneNumbers = append(phoneNumbers, user.Phone)
		}(id)
	}

	wg.Wait()
	fmt.Println(7, phoneNumbers)

	for _, id := range input.IDs {
		wg.Add(1)

		go func(id int) {
			defer wg.Done()
			var user UserSchema
			if err := db3.Where("id = ?", id).First(&user).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			mu.Lock()
			defer mu.Unlock()

			user.Mentor = input.MentorName
			if err := db2.Save(&user).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}(id)
	}
	wg.Wait()

	for _, phone := range phoneNumbers {
		wg.Add(1)

		go func(phone string) {
			defer wg.Done()
			var reuser RenrollSchema
			if err := db2.Where("phone = ?", phone).First(&reuser).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			mu.Lock()
			defer mu.Unlock()

			reuser.Mentor = input.MentorName
			if err := db2.Save(&reuser).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}(phone)
	}
	wg.Wait()

	c.JSON(http.StatusOK, gin.H{"message": "Mentor and students updated successfully"})
}

func DelMentorGet(c *gin.Context) {
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
		if mentor.Onn == 0 {
			wg1.Add(1)
			go func(mentor MentorSchema) {
				defer wg1.Done()
				mentorsChan <- mentor
			}(mentor)
		}

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

func DelMentor(c *gin.Context) {
	var input deleteMentSchema
	if err := c.ShouldBindBodyWithJSON(&input); err != nil {
		fmt.Println("2")
		fmt.Println(input)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	fmt.Println("1")
	fmt.Println(input)

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
			result := db2.Where("id=?", id).Delete(&MentorSchema{})

			// Check for errors
			if err := result.Error; err != nil {
				// Handle error (e.g., log it)
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
		}(id)
	}
	// Wait for all Goroutines to finish
	wg.Wait()

	c.JSON(http.StatusOK, gin.H{"message": "Mentor data deleted successfully"})
}
