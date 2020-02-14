package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strings"
	"time"
)

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

func writeMessageToFile(msg Message, pathToFile string) error {
	f, err := os.OpenFile(pathToFile,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.WriteString("\n"+fmt.Sprint(msg)); err != nil {
		return err
	}
	return nil
}

func readMessageFromFileById(id [16]byte, index *DBPosIndex, pathToFile string) Message{
	f,err := os.Open(pathToFile)
	check(err)
	defer f.Close()

	pos, ok := (*index)[id]
	if !ok {
		log.Println("the message id doesn't exist in the index map")
		return Message{}
	}
	offset:= pos // int64(lenHeader)+pos
	var whence int = 0
	_, err = f.Seek(offset, whence)
	check(err)

	r := csv.NewReader(f)
	record, err := r.Read()
	if err == io.EOF {
		log.Println("EOF reading message from file")
		return Message{}
	}
	check(err)

	form := "2006-01-02T15:04:05-07:00"
	t, err := time.Parse(form, record[4])
	check(err)

	msg := Message {
		Id: 			record[0],
		Name: 			record[1],
		Email:			record[2],
		Text: 			record[3],
		CreationTime: 	t,
	}
	return msg
}
