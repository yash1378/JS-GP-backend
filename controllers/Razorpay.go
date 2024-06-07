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
	"github.com/razorpay/razorpay-go/utils"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type order struct {
	Amount int `json:"amount"`
}

type veri struct {
	Payid   string `json:"razorpay_payment_id"`
	Orderid string `json:"razorpay_order_id"`
	Sign    string `json:"razorpay_signature"`
}

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

func Order(c *gin.Context) {
	var inp map[string]interface{}

	if err := c.ShouldBindJSON(&inp); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if "payload" exists and is not nil
	payload, ok := inp["payload"].(map[string]interface{})
	if !ok {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "payload is missing or invalid"})
		return
	}

	// Extract "order" from "payload" if it exists
	order, ok := payload["payment"].(map[string]interface{})
	if !ok {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "order is missing or invalid"})
		return
	}

	entity, ok := order["entity"].(map[string]interface{})
	if !ok {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "entity is missing or invalid in order"})
		return
	}

	orderNotes, ok := entity["notes"].(map[string]interface{})
	if !ok {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "notes is missing or invalid in order entity"})
		return
	}
	program := orderNotes["program"].(string)
	if (program!="Premium" && program!="Normal") {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "entity is missing or invalid in order"})
		return
	}
	// Now you can access individual fields within the "notes" object
	class := orderNotes["class"].(string)
	email := orderNotes["email"].(string)
	name := orderNotes["name"].(string)
	phone := orderNotes["phone"].(string)


	fmt.Println("executed1")
	fmt.Println(program)

	var wg sync.WaitGroup
	resultChan := make(chan int, 1)
	wg.Add(1)

	var input User
	input.Email = email
	input.Name = name
	input.Phone = phone
	input.Class = class



	fmt.Println("executed2")

	if program=="Normal" {
		input.Sub = "Normal"
	} else {
		input.Sub = "Premium"
	}


	// return

	// Assigning today's date to input.Date in the format YYYY-MM-DD
	today := time.Now()
	input.Date = today.Format("2006-01-02") // Using the layout format "2006-01-02" for YYYY-MM-DD

	fmt.Println("executed3")

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

	fmt.Println("executed4")
	fmt.Println(input)

	// Wait for the result from the goroutine
	result := <-resultChan
	fmt.Println(result)
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

	fmt.Println("executed5")

	reuser := models.RenrollSchema{Name: input.Name, Phone: input.Phone, Email: input.Email, Date: input.Date, Class: input.Class, Sub: input.Sub, Mentor: "", Renrollment: 0}
	if err := db1.Create(&reuser).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to save data"})
		return
	}

	fmt.Println("executed6")

	fmt.Println(class)
	fmt.Println(name)
	fmt.Println(email)
	fmt.Println(phone)
	fmt.Println(today)

	c.JSON(http.StatusOK, gin.H{"message": "User data saved successfully"})
}

func Key(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"key": "rzp_test_gAnipV9pamMXV2"})
}

func Verify(c *gin.Context) {
	var input veri

	if err := c.ShouldBindBodyWithJSON(&input); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// client := razorpay.NewClient("rzp_test_gAnipV9pamMXV2", "cUhcdokPA3QwiULcMmadGx86")

	params := map[string]interface{}{
		"razorpay_order_id":   input.Orderid,
		"razorpay_payment_id": input.Payid,
	}

	signature := input.Sign
	secret := "cUhcdokPA3QwiULcMmadGx86"
	utils.VerifyPaymentSignature(params, signature, secret)

	redirectURL := fmt.Sprintf("http://localhost:3000/razor?reference=%v", input.Payid)
	c.Redirect(http.StatusFound, redirectURL)

	// c.JSON(http.StatusOK, gin.H{"message": "payment success"})
}
