package models

import (
	"log"
	"os"

	"github.com/joho/godotenv"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB1 *gorm.DB

type UserSchema struct {
	ID     uint   `json:"id" gorm:"primaryKey"`
	Name   string `json:"name"`
	Phone  string `json:"phone"`
	Email  string `json:"email"`
	Date   string `json:"date"`
	Class  string `json:"class"`
	Sub    string `json:"sub"`
	Mentor string `json:"mentorName"`
}

type MentorSchema struct {
	ID      uint   `json:"id" gorm:"primaryKey"`
	Name    string `json:"name"`
	College string `json:"college"`
	Date    string `json:"date"`
	Phone   string `json:"phone"`
	Handle  int    `json:"-" gorm:"default:0"`
	Onn     int    `json:"-" gorm:"default:0"`
	Total   int    `json:"-" gorm:"default:0"`
}

type MentorLogin struct {
	ID         uint   `json:"id" gorm:"primaryKey"`
	Email      string `json:"email"`
	Password   string `json:"-" gorm:"default:"`
	MentorName string `json:"name"`
}

type RenrollSchema struct {
	ID          uint   `json:"id" gorm:"primaryKey"`
	Name        string `json:"name"`
	Phone       string `json:"phone"`
	Email       string `json:"email"`
	Date        string `json:"date"`
	Class       string `json:"class"`
	Sub         string `json:"sub"`
	Mentor      string `json:"-" gorm:"default:"`
	Renrollment uint   `json:"-" gorm:"default:0"`
}

type OwnerSchema struct {
	ID        uint   `json:"id" gorm:"primaryKey"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	OwnerName string `json:"ownername"`
}

func init() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file")
	}
}

func ConnectDatabase() {

	host := os.Getenv("DB_HOST")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	port := os.Getenv("DB_PORT")

	dsn := "host=" + host + " user=" + user + " password=" + password + " dbname=" + dbname + " port=" + port
	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{}) // change the database provider if necessary

	if err != nil {
		panic("Failed to connect to database!")
	}

	// Check and create UserSchema table
	if !database.Migrator().HasTable(&UserSchema{}) {
		database.AutoMigrate(&UserSchema{})
	}

	// Check and create MentorSchema table
	if !database.Migrator().HasTable(&MentorSchema{}) {
		database.AutoMigrate(&MentorSchema{})
	}

	if !database.Migrator().HasTable(&RenrollSchema{}) {
		database.AutoMigrate(&RenrollSchema{})
	}

	if !database.Migrator().HasTable(&MentorLogin{}) {
		database.AutoMigrate(&MentorLogin{})
	}

	if !database.Migrator().HasTable(&OwnerSchema{}) {
		database.AutoMigrate(&OwnerSchema{})
	}
	DB1 = database

}
