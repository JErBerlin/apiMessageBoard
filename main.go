package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type Message struct {
	Id				string
	Name			string
	Email			string
	Text			string
	CreationTime	time.Time
}

// PrintFields return field names as they appear in the type definition
func (m Message) PrintFields() string {

	var fields []string
	for i:=0; i < reflect.TypeOf(m).NumField(); i++ {
		fields = append(fields,
			strings.ToLower(reflect.TypeOf(m).Field(i).Name))
	}
	return strings.Join(fields, " ")
}

func validHeaderRow(row []string) (bool, error) {
	// remove all non alphanumeric chars from headers
	reg, err := regexp.Compile("[^a-zA-Z0-9]+")
	check(err)
	var ws []string
	for _, w := range row {
		ws = append(ws, strings.ToLower(reg.ReplaceAllString(w, "")))
	}
	headerRowHave := strings.Join(ws, " ")
	headerRowWant := fmt.Sprintf(Message{}.PrintFields())
	if headerRowWant != headerRowHave {
		err := errors.New("error parsing csv file: header row is invalid")
		return false, err
	}
	return true, nil
}

func readMessagesFromFile(pathToFile string) []Message {
	messages := make([]Message, 0)

	// Open the file
	csvFile, err := os.Open(pathToFile)
	check(err)
	defer csvFile.Close()

	// Parse the file
	r := csv.NewReader(csvFile)

	// suppose first line is header make sure it is valid
	record, err := r.Read()
	check(err)
	_, err = validHeaderRow(record)
	check(err)

	// Iterate the message records
	for {
		// Read each record from csv
		record, err = r.Read()
		if err == io.EOF {
			break
		}
		check(err)

		// TODO format creation time as separate func
		form := "2006-01-02T15:04:05-07:00"
		t, err := time.Parse(form, record[4])
		check(err)

		newMessage := Message {
			Id: 			record[0],
			Name: 			record[1],
			Email:			record[2],
			Text: 			record[3],
			CreationTime: 	t,
		}
		messages = append(messages, newMessage)
	}

	return messages
}

func main() {
	messages := readMessagesFromFile("./messages.csv")
	// message := messages[0]

	startServingMessages(messages)
}

func startServingMessages(msg interface{}) {
	r := gin.Default()
	r.GET("/messages", func(c *gin.Context){
		c.JSON(200, msg)
	})
	r.Run() // listen and serve on 0.0.0.0:8080 ("localhost:8080")
}
func check(e error) {
	if e != nil {
		log.Fatal(e)
	}
}