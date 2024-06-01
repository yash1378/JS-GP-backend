package main

import (
	"guidance/controllers"
	"guidance/models"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	config := cors.Config{
		AllowOrigins:     []string{"*"}, // Allow from specific origin, use "*" to allow all
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
	r := gin.Default()

	// Use the CORS middleware with the configured settings
	r.Use(cors.New(config))
	models.ConnectDatabase()
	r.GET("/data_without_mentor", controllers.UserWithoutMentor)
	r.GET("/api/data", controllers.UserGet)                      //checked and Final
	r.POST("/", controllers.UserPost)                            //checked and Final
	r.GET("/api/mentorData", controllers.MentorGet)              //checked and Final
	r.POST("/mentorData", controllers.MentorPost)                //checked and Final
	r.POST("/student/:phone", controllers.MentorUpdate)          //checked and Final
	r.POST("/api/update", controllers.MentorStudentUpdate)       //checked and Final
	r.POST("/mentorupdate/:phone", controllers.MentorInfoUpdate) //checked
	r.GET("/api/renrollData", controllers.RenrollDataGet)        //checked and Final
	r.POST("/renrollment", controllers.RenrollDataPost)          //checked and Final
	r.POST("/change-password", controllers.ChangePassword)       //checked and Final
	r.GET("/api/ownerData", controllers.OwnerGet)                //checked and Final
	r.POST("/ownerData", controllers.OwnerPost)                  //checked and Final
	r.DELETE("/api/delete", controllers.DELETE)                  //checked and Final
	r.POST("/api/finalMentor", controllers.FinalMentor)          //checked and Final
	r.POST("/order", controllers.Order)
	r.GET("/getkey", controllers.Key)
	r.POST("/api/paymentverify", controllers.Verify)
	// Start the server
	r.Run(":8080")
}
