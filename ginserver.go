package main

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"sort"
	"time"
)

//TODO: refactor1 some the decode, write and read functionality out the of the gin server file to separate purposes
//TODO: refactor2 some of the repeated code in different handles that does the same

// startServingMessages serves the whole list of messages, body response is in JSON format
// (only suitable for small messages files)
func startRouter() {
	r := gin.Default()

	// for the private API: see one messages, list all, edit
	authorized := r.Group("/", gin.BasicAuth(gin.Accounts{
		"admin": "back-challenge",
	}))

	r.GET("/", getHomePage)
	r.POST("/new", postMessage)
	authorized.GET("/view/:id", viewOneMessageByPath)
	authorized.POST("/edit", editMessage)
	authorized.GET("/messages", getAllMessages)  // not functional yet, should be streaming for big files

	r.Run() // listen and serve on 0.0.0.0:8080 ("localhost:8080")
}

func getHomePage(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Welcome to the back message board. You can POST a new entry at /new," +
		"GET one message by id at /view/:id, " +
		"edit (POST) a message at /edit or also " +
		"GET all messages at /messages. The last three require authorization."})
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

// editMessage is a handler for editing messages by admin using id
// the only field that can be modified is text
// the request has to pass a JSON with a valid existing id and a new text
func editMessage(c *gin.Context) {
	body := c.Request.Body
	value, err := ioutil.ReadAll(body)
	if err!= nil {
		log.Println("body of the request could not be read, ", err)
		c.JSON(http.StatusUnprocessableEntity , gin.H{"message": "the post request is invalid"})
		return
	}
	newMessage := Message{}
	json.Unmarshal(value, &newMessage)
	idStr := newMessage.Id
	if idStr == "" {
		log.Println("unmarshalling of the json message to be edited failed, ", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "the id of the message to be edited doesn't exist or is bad formatted"})
		return
	}

	// refresh indexation before search of editing message
	pathToFile := PathToMessagesFile
	mapIdPos, err := fillPositionIndex(pathToFile)
	if err != nil {
		log.Println("indexing of messages failed during edit request process, ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "the resource requested cannot be edited"})
		return
	}

	id := idToHex16byte(idStr)
	if _, ok := (*mapIdPos)[id]; !ok {
		log.Println("the id of the message to be edited cannot be found in the index")
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "the id of the message to be edited could not be found"})
		return
	}
	oldMessage := readMessageFromFileById(id, mapIdPos, pathToFile)
	// copy relevant fields (name, from old message to new
	newMessage.Name = oldMessage.Name
	newMessage.Email = oldMessage.Email
	newMessage.CreationTime = oldMessage.CreationTime

	err = replaceMessageInFileById(newMessage, id, mapIdPos, PathToMessagesFile)
	if err!= nil {
		log.Println("replace operation of the in-file message failed ")
		c.JSON(http.StatusInternalServerError , gin.H{"message": "the post requested could not be made"})
		return
	}
	c.JSON(http.StatusOK , gin.H{"message": "message was edited successfully"})
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

// getAllMessages retrieve all messages from file, sorts them anti-chronologically and send them back 
// (as a unique response from a RESTful API?)
func getAllMessages (c *gin.Context) {
	pathToFile := PathToMessagesFile

	// do indexing again
	mapIdPos, err := fillPositionIndex(pathToFile)
	if err != nil {
		log.Println("indexing of messages failed during request process, ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "the ressource requested cannot be served"})
		return
	}
	dbChronFinder, err := fillChronIndArr(pathToFile)
	if err != nil {
		log.Println("indexing of messages failed during request process, ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "the ressource requested cannot be served"})
		return
	}

	// remove repeated instances in TimeArr
	dbChronFinder.TimeArr = appendUniques(*(dbChronFinder.TimeArr), *(dbChronFinder.TimeArr))

	// sort the times anti-chronologically
	sort.Slice(*(dbChronFinder.TimeArr), func(i, j int) bool{ 
		return (*dbChronFinder.TimeArr)[i] > (*dbChronFinder.TimeArr)[j] })


	messages := make([]Message,0, len(*dbChronFinder.TimeArr))
	// TODO: Debuging -- const testLen
	for i:=0; i < len(*dbChronFinder.TimeArr); i++ {
		t := (*dbChronFinder.TimeArr)[i]
		id := (*dbChronFinder.ChronIndex)[t]

		oneMessage := readMessageFromFileById(id, mapIdPos, pathToFile)
		// DEBUG start
		log.Printf("%4d: \t%v -- %s\n", i+1, oneMessage.CreationTime.Format("02/01/2006- 15:04:05"), oneMessage.Id)
		// DEBUG end
		messages = append(messages,oneMessage)
	}
	c.JSON(http.StatusOK, messages)
}

func appendUniques(a []int64, b []int64) *[]int64 {
	check := make(map[int64]int)
	d := append(a, b...)
	res := make([]int64,0)
	for _, val := range d {
		check[val] = 1
	}
	for num, _ := range check {
		res = append(res,num)
	}

	return &res
}


