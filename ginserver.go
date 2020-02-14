package main

import "github.com/gin-gonic/gin"

// startServingMessages serves the whole list of messages, body response is in JSON format
// (only suitable for small messages files)
func startServingMessages(msg interface{}) {
	r := gin.Default()
	r.GET("/messages", func(c *gin.Context){
		c.JSON(200, msg)
	})
	r.Run() // listen and serve on 0.0.0.0:8080 ("localhost:8080")
}
