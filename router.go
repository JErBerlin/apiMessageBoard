// router.go does the routing of path to handlers and define the handler functions
package main

import (
	"encoding/json"
	"fmt"
	"github.com/JErBerlin/back_message_board/db"
	"github.com/JErBerlin/back_message_board/message"
	"github.com/gin-gonic/gin"
	"gopkg.in/olahol/melody.v1"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"strings"
)

const port = ":8080"

// TODO: refactor1 some the decode, write and read functionality out the of the gin server file to separate purposes
// TODO: in startRouter: separate functionality not related to routing

func StartRouter() {
	// default gin router (logger and recovery functions)
	r := gin.Default()
	authorized := r.Group("/", gin.BasicAuth(gin.Accounts{
		"admin": "back-challenge",
	}))

	// public API
	// new message
	r.POST("/new", postMessage)

	// private API
	// list one message, edit
	authorized.GET("/view/:id", viewOneMessageByPath)
	authorized.POST("/edit", editMessage)

	// websocket API
	// list all messages, list one message

	// new instance for websocket communication (just for admin)
	m := melody.New()
	m.Upgrader.CheckOrigin = func(r *http.Request) bool { return true }

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

			// do indexing again
			mapIdPos, err := db.FillPositionIndex(pathToMessagesFile)
			if err != nil {
				log.Println("indexing of messages failed during websocket communication, ", err)
				m.Broadcast([]byte("the ressource requested cannot be served\n"))
				return
			}
			dbChronFinder, err := db.FillChronIndArr(pathToMessagesFile)
			if err != nil {
				log.Println("indexing of messages failed during websocket communication ", err)
				m.Broadcast([]byte("the ressource requested cannot be served\n"))
				return
			}

			// remove repeated instances in TimeArr
			dbChronFinder.TimeArr = appendUniques(*(dbChronFinder.TimeArr), *(dbChronFinder.TimeArr))

			// sort the times anti-chronologically
			sort.Slice(*(dbChronFinder.TimeArr), func(i, j int) bool {
				return (*dbChronFinder.TimeArr)[i] > (*dbChronFinder.TimeArr)[j]
			})

			for i := 0; i < len(*dbChronFinder.TimeArr); i++ {
				t := (*dbChronFinder.TimeArr)[i]
				id := (*dbChronFinder.ChronIndex)[t]

				record, err := db.ReadMessageFromFileById(id, mapIdPos, pathToMessagesFile)
				if err != nil {
					log.Println(err)
					return
				}
				oneMessage, err := message.NewFromRecord(record)
				if err != nil {
					log.Println(err)
					return
				}
				data, err := json.Marshal(oneMessage)
				if err != nil {
					fmt.Println("failed json encoding in websocket communication")
					m.Broadcast([]byte("the ressource requested cannot be served\n"))
					return
				}
				m.Broadcast([]byte(string(data) + "\n"))
			}
			m.Broadcast([]byte("-- end of the data streaming --\n"))
		} else {
			mapIdPos, err := db.FillPositionIndex(pathToMessagesFile)
			if err != nil {
				fmt.Println("failed indexing in websocket communication")
				m.Broadcast([]byte("-- could not serve the required resource --\n"))
				return
			}
			id, _ := message.IdToHex16byte(string(msg[2:]))
			if _, ok := (*mapIdPos)[id]; ok {
				record, err := db.ReadMessageFromFileById(id, mapIdPos, pathToMessagesFile)
				if err != nil {
					log.Println(err)
					return
				}
				oneMessage, err := message.NewFromRecord(record)
				if err != nil {
					log.Println(err)
					return
				}
				data, err := json.Marshal(oneMessage)
				if err != nil {
					fmt.Println("failed json encoding in websocket communication")
					return
				}
				m.Broadcast([]byte(string(data) + "\n"))
			} else {
				m.Broadcast([]byte("-- input command not recognized, or the id doesn't exist --\n"))
			}
		}
	})

	r.Run(port) // listen and serve on 0.0.0.0:port ("localhost:port")
}

func postMessage(c *gin.Context) {
	body := c.Request.Body
	msgJSON, err := ioutil.ReadAll(body)
	if err != nil {
		log.Println("body from post request could not be read:", err)
		c.JSON(http.StatusUnprocessableEntity, gin.H{"message": "the post request is invalid"})
		return
	}
	newMessage, err := message.NewFromJSON(msgJSON)
	if err != nil {
		log.Println("new message could not be generated from json message:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "the post requested could not be made"})
		return
	}

	err = db.WriteMessageToFile(newMessage, pathToMessagesFile)
	if err != nil {
		log.Println("new message could not be written:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "the post requested could not be made"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "message was posted successfully"})
}

// editMessage is a handler for editing messages by admin using id
// the only field that can be modified is text
// the request has to pass a JSON with a valid existing id and a new text
func editMessage(c *gin.Context) {
	body := c.Request.Body
	value, err := ioutil.ReadAll(body)
	if err != nil {
		log.Println("body of the request could not be read, ", err)
		c.JSON(http.StatusUnprocessableEntity, gin.H{"message": "the post request is invalid"})
		return
	}
	newMessage := message.Message{}
	json.Unmarshal(value, &newMessage)
	idStr := newMessage.Id
	if idStr == "" {
		log.Println("unmarshalling of the json message to be edited failed, ", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "the id of the message to be edited is not set or is bad formatted"})
		return
	}

	mapIdPos, err := db.FillPositionIndex(pathToMessagesFile)
	if err != nil {
		log.Println("indexing of messages failed during request process, ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "the edition of the message could not be made"})
		return
	}
	id, _ := message.IdToHex16byte(idStr)
	err = db.ReplaceMessageInFileById(newMessage, id, mapIdPos, pathToMessagesFile)
	if err != nil {
		log.Println("old message could not be replaced by new message: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "the edition of the message could not be made"})
	}
	c.JSON(http.StatusOK, gin.H{"message": "message was edited successfully"})
}

func viewOneMessageByPath(c *gin.Context) {
	mapIdPos, err := db.FillPositionIndex(pathToMessagesFile)
	if err != nil {
		log.Println("indexing of messages failed during request process, ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "the ressource requested cannot be served"})
		return
	}

	idStr := c.Param("id")
	if idStr == "" {
		log.Println("the id of the message requested doesn't exist or is bad formatted")
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "the id of the message requested doesn't exist or is bad formatted"})
	}
	id, _ := message.IdToHex16byte(idStr)
	record, err := db.ReadMessageFromFileById(id, mapIdPos, pathToMessagesFile)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "the message requested could not be found"})
		return
	}
	oneMessage, err := message.NewFromRecord(record)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "the message requested could not be correctly processed"})
		return
	}
	c.JSON(http.StatusOK, oneMessage)
}

func appendUniques(a []int64, b []int64) *[]int64 {
	check := make(map[int64]int)
	d := append(a, b...)
	res := make([]int64, 0)
	for _, val := range d {
		check[val] = 1
	}
	for num, _ := range check {
		res = append(res, num)
	}

	return &res
}
