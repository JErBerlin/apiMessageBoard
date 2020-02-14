package main

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"time"
)

//TODO: refactor some the decode, write and read functionality out the of the gin server file to separate purposes
func getHomePage(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Welcome to the back message board. You can authenticate at /admin, " +
		"query one message by id at /view/:id, " +
		"post a new entry at /new or also " +
		"retrieve all messages at /messages."})
}

func postMessage(c *gin.Context) {
	body := c.Request.Body
	value, err := ioutil.ReadAll(body)
	if err!= nil {
		c.JSON(http.StatusUnprocessableEntity , gin.H{"message": "the post request is invalid"})
		return
	}
	oneMessage := Message{}
	json.Unmarshal(value, &oneMessage)

	newTime := time.Now()
	newSource := rand.NewSource(int64(newTime.Nanosecond()))
	randNow := rand.New(newSource)
	newIdStr, _ := randomIdStr16(randNow)
	oneMessage.Id = newIdStr
	oneMessage.CreationTime = newTime

	err = writeMessageToFile(oneMessage, PathToMessagesFile)
	if err!= nil {
		c.JSON(http.StatusInternalServerError , gin.H{"message": "the post requested could not be made"})
		return
	}
	c.JSON(http.StatusOK , gin.H{"message": "message was posted successfully"})
}

//TODO: do we need query method apart from view message by path?
func messageQuery (c *gin.Context) {
	pathToFile := PathToMessagesFile
	mapIdPos, err := fillPositionIndex(pathToFile)
	if err != nil {
		log.Println("indexing of messages failed during request process, ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "the ressource requested cannot be served"})
		return
	}

	idStr := c.Query("id")
	if idStr != "" {
		id := idToHex16byte(idStr)
		if _, ok := (*mapIdPos)[id]; ok {
			oneMessage := readMessageFromFileById(id, mapIdPos, pathToFile)
			c.JSON(http.StatusOK, oneMessage)
			return
		}
	}
	c.JSON(http.StatusBadRequest, gin.H{"message": "the id of the message requested doesn't exist or is bad formatted"})
}

func viewOneMessageByPath (c *gin.Context) {
	pathToFile := PathToMessagesFile
	mapIdPos, err := fillPositionIndex(pathToFile)
	if err != nil {
		log.Println("indexing of messages failed during request process, ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "the ressource requested cannot be served"})
		return
	}

	idStr := c.Param("id")
	if idStr != "" {
		id := idToHex16byte(idStr)
		if _, ok := (*mapIdPos)[id]; ok {
			oneMessage := readMessageFromFileById(id, mapIdPos, pathToFile)
			c.JSON(http.StatusOK, oneMessage)
			return
		}
	}
	c.JSON(http.StatusBadRequest, gin.H{"message": "the id of the message requested doesn't exist or is bad formatted"})
}

// startServingMessages serves the whole list of messages, body response is in JSON format
// (only suitable for small messages files)
func startServing() {
	r := gin.Default()
	r.GET("/", getHomePage)
	r.GET("/query", messageQuery)	// could have different functionality than view
	r.GET("/view/:id", viewOneMessageByPath)
	r.POST("/new", postMessage)
	r.GET("/messages", getHomePage)  // not functional yet, should be streaming for big files
	r.Run() // listen and serve on 0.0.0.0:8080 ("localhost:8080")
}
