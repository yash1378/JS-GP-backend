package main

import (
	"guidance/controllers"

	"github.com/gin-gonic/gin"
)

func main() {

	r := gin.Default()
	// models.ConnectDatabase()
	r.GET("/api/data", controllers.UserGet)
	r.POST("/", controllers.UserPost)
	r.GET("/api/mentorData", controllers.MentorGet)
	r.POST("/mentorData", controllers.MentorPost)
	// r.Static("/static", "./static")

	// // Route for serving the index.html file
	// r.GET("/", func(c *gin.Context) {
	// 	c.File("./static/index.html")
	// })

	// Start the server
	r.Run(":8080")
}
