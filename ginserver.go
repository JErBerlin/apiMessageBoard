package main

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"gopkg.in/olahol/melody.v1"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"sort"
	"strings"
	"time"
)

const port = ":8080"

// TODO: refactor1 some the decode, write and read functionality out the of the gin server file to separate purposes
// TODO: refactor2 some of the repeated code in different handles that does the same
// TODO: in startRouter: separate functionality not related to routing

func startRouter() {
	// default gin router (logger and recovery functions)
	r := gin.Default()

	// new instance for websocket communication (just for admin)
	m := melody.New()
	m.Upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	// for the private API: see one messages, list all, edit
	authorized := r.Group("/", gin.BasicAuth(gin.Accounts{
		"admin": "back-challenge",
	}))

	// public API
	r.POST("/new", postMessage)

	// private API
	authorized.GET("/view/:id", viewOneMessageByPath)
	authorized.POST("/edit", editMessage)
	//authorized.GET("/messages", getAllMessages)  // not functional yet, not suited for big csv file

	// websocket: access via web application
	authorized.GET("/", func(c *gin.Context) {
		http.ServeFile(c.Writer, c.Request, "index.html")
	})
	authorized.GET("/ws", func(c *gin.Context) {
		m.HandleRequest(c.Writer, c.Request)
	})
	m.HandleMessage(func(s *melody.Session, msg []byte) {
		m.Broadcast(msg)
		if strings.ToLower(string(msg)) == "> all" {
			m.Broadcast([]byte("-- retrieving all messages in streaming... --\n"))

			pathToFile := PathToMessagesFile

			// do indexing again
			mapIdPos, err := fillPositionIndex(pathToFile)
			if err != nil {
				log.Println("indexing of messages failed during websocket communication, ", err)
				m.Broadcast([]byte("the ressource requested cannot be served\n"))
				return
			}
			dbChronFinder, err := fillChronIndArr(pathToFile)
			if err != nil {
				log.Println("indexing of messages failed during websocket communication ", err)
				m.Broadcast([]byte("the ressource requested cannot be served\n"))
				return
			}

			// remove repeated instances in TimeArr
			dbChronFinder.TimeArr = appendUniques(*(dbChronFinder.TimeArr), *(dbChronFinder.TimeArr))

			// sort the times anti-chronologically
			sort.Slice(*(dbChronFinder.TimeArr), func(i, j int) bool{
				return (*dbChronFinder.TimeArr)[i] > (*dbChronFinder.TimeArr)[j] })

			for i:=0; i < len(*dbChronFinder.TimeArr); i++ {
				t := (*dbChronFinder.TimeArr)[i]
				id := (*dbChronFinder.ChronIndex)[t]

				oneMessage := readMessageFromFileById(id, mapIdPos, pathToFile)
				data, err := json.Marshal(oneMessage)
				if err != nil {
					fmt.Println("failed json encoding in websocket communication")
					m.Broadcast([]byte("the ressource requested cannot be served\n"))
					return
				}
				m.Broadcast([]byte(string(data)+"\n"))
			}
			m.Broadcast([]byte("-- end of the data streaming --\n"))
		} else {
			mapIdPos, err := fillPositionIndex(PathToMessagesFile)
			if err != nil {
				fmt.Println("failed indexing in websocket communication")
				m.Broadcast([]byte("-- could not serve the required resource --\n"))
				return
			}
			id := idToHex16byte(string(msg[2:]))
			if _, ok := (*mapIdPos)[id]; ok {
				oneMessage := readMessageFromFileById(id, mapIdPos, PathToMessagesFile)
				data, err := json.Marshal(oneMessage)
				if err != nil {
					fmt.Println("failed json encoding in websocket communication")
					return
				}
				m.Broadcast([]byte(string(data)+"\n"))
			} else {
				m.Broadcast([]byte("-- input command not recognized, or the id doesn't exist --\n"))
			}
		}
	})

	r.Run(port) // listen and serve on 0.0.0.0:port ("localhost:port")
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


