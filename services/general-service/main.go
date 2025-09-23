package main

import "github.com/gin-gonic/gin"

func main () {
	var server = gin.Default()

	server.GET("/ping", func (context *gin.Context) {
		context.JSON(200, gin.H{
			"message": "pong",
		})
	})
	
	server.Run(":8080")
}