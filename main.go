package main

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"os"
	"reflect"
	"regexp"
	"strings"
)

type Message struct {
	Id				string
	Name			string
	Email			string
	Text			string
	CreationTime	string // time.Time
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
		if err != nil {
			log.Fatal(err)
		}

		newMessage := Message {
			Id: 			record[0],
			Name: 			record[1],
			Email:			record[2],
			Text: 			record[3],
			CreationTime: 	record[4],
		}
		messages = append(messages, newMessage)
	}

	return messages
}

func main() {
	messages := readMessagesFromFile("./messages.csv")

	fmt.Printf("First message, creation time: %v\n", messages[0].CreationTime)

	messageJSON, err := json.Marshal(messages[0])
	check(err)

	fmt.Printf("First message JSON: %s\n", string(messageJSON))
	// startServing()
}

func startServing() {
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context){
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	r.Run() // listen and serve on 0.0.0.0:8080 ("localhost:8080")
}
func check(e error) {
	if e != nil {
		log.Fatal(e)
	}
}