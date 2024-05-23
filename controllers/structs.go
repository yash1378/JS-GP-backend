package controllers

import (
	"gorm.io/gorm"
)

var db1 *gorm.DB //for UserData
var db2 *gorm.DB //for MentorData

// MentorData.go
type Mentor struct {
	Name    string `json:"name" binding:"required"`
	College string `json:"college" binding:"required"`
	Date    string `json:"date" binding:"required"`
	Handle  int    `json:"-"`
	Onn     int    `json:"-"`
	Total   int    `json:"-"`
}

type MentorSchema struct {
	ID      uint   `json:"id"`
	Name    string `json:"name"`
	College string `json:"college"`
	Date    string `json:"date"`
	Handle  int    `json:"-"`
	Onn     int    `json:"-"`
	Total   int    `json:"-"`
}

// UserData.go
type User struct {
	Name   string `json:"name" binding:"required"`
	Phone  string `json:"phone" binding:"required"`
	Email  string `json:"email" binding:"required"`
	Date   string `json:"date" binding:"required"`
	Class  string `json:"class" binding:"required"`
	Sub    string `json:"sub" binding:"required"`
	Mentor string `json:"mentorName" binding:"required"`
}

type UserSchema struct {
	ID     uint   `json:"id"`
	Name   string `json:"name"`
	Phone  string `json:"phone"`
	Email  string `json:"email"`
	Date   string `json:"date"`
	Class  string `json:"class"`
	Sub    string `json:"sub"`
	Mentor string `json:"mentorName"`
}
