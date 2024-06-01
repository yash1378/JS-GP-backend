package controllers

import (
	"gorm.io/gorm"
)

var db1 *gorm.DB //for UserData
var db2 *gorm.DB //for MentorData
var db3 *gorm.DB //for StudentData
var db4 *gorm.DB //for RenrollData
var db5 *gorm.DB //for OwnerData

type MentorLogin struct {
	ID         uint   `json:"id"`
	Email      string `json:"email"`
	Password   string `json:"password"`
	MentorName string `json:"name"`
}

// MentorData.go
type Mentor struct {
	Name     string `json:"name" binding:"required"`
	Phone    string `json:"phone" binding:"required"`
	Date     string `json:"date" binding:"required"`
	College  string `json:"college" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type FinalMentorSchema struct {
	IDs        []int  `json:"studentIds"`
	MentorName string `json:"mentorName"`
}

type MentorSchema struct {
	ID      uint   `json:"id"`
	Name    string `json:"name"`
	College string `json:"college"`
	Phone   string `json:"phone"`
	Date    string `json:"date"`
	Handle  int    `json:"handle"`
	Onn     int    `json:"on"`
	Total   int    `json:"total"`
}

type UpdateCount struct {
	MentorName   string `json:"mentorName" binding:"required"`
	StudentCount int    `json:"studentCount" binding:"required"`
}

type MentorUpdateInfo struct {
	Name    string `json:"name" binding:"required"`
	Email   string `json:"email" binding:"required"`
	PhoneNo string `json:"phoneno" binding:"required"`
}

type PasswordChange struct {
	Email       string `json:"email" binding:"required"`
	OldPassword string `json:"oldPassword" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required"`
}

// UserData.go
type User struct {
	Name   string `json:"name" binding:"required"`
	Phone  string `json:"phone" binding:"required"`
	Email  string `json:"email" binding:"required"`
	Date   string `json:"date" binding:"required"`
	Class  string `json:"class" binding:"required"`
	Sub    string `json:"sub" binding:"required"`
	Mentor string `json:"mentorName"`
}

type UserSchema struct {
	ID     uint   `json:"id"`
	Name   string `json:"name"`
	Phone  string `json:"phone"`
	Email  string `json:"email"`
	Date   string `json:"date"`
	Class  string `json:"class"`
	Sub    string `json:"sub"`
	Mentor string `json:"mentor"`
}

type deleteSchema struct {
	IDs     []uint   `json:"ids"`
	Mentors []string `json:"mentors"`
}

// for StudentData
type Student struct {
	Name      string `json:"studentName" binding:"required"`
	Phone     string `json:"phoneNumber" binding:"required"`
	Email     string `json:"studentEmail" binding:"required"`
	Class     string `json:"selectedClass" binding:"required"`
	Date      string `json:"selectedDate" binding:"required"`
	NewMentor string `json:"newmentor" binding:"required"`
}

type RenrollSchema struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Phone       string `json:"phone"`
	Email       string `json:"email"`
	Date        string `json:"date"`
	Class       string `json:"class"`
	Sub         string `json:"sub"`
	Mentor      string `json:"mentor"`
	Renrollment uint   `json:"renrollment"`
}
type Renroll struct {
	Name        string `json:"studentName"`
	Phone       string `json:"phone"`
	Email       string `json:"email"`
	Date        string `json:"date"`
	Class       string `json:"classs"`
	Sub         string `json:"sub"`
	Mentor      string `json:"mentorName"`
	Renrollment uint   `json:"-"`
}

func (UserSchema) TableName() string {
	return "user_schemas"
}
func (MentorSchema) TableName() string {
	return "mentor_schemas"
}

type Owner struct {
	Email     string `json:"email" binding:"required"`
	Password  string `json:"password" binding:"required"`
	OwnerName string `json:"ownername" binding:"required"`
}

type OwnerSchema struct {
	ID        uint   `json:"id"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	OwnerName string `json:"ownername"`
}
