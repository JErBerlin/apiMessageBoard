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

// replaceMessageInFileById leaves the old message unchanged and writes another one with the same id and the same
// fields, but with a new text, at the end of the file
// this way of overwriting messages by appending works since the index system only remembers the last record (line) of
// a group of records with identical ids
// TODO: think of a possible garbage collector or compactifier to get rid of repeated ids resulting from edited messages
func replaceMessageInFileById(msg Message, id [16]byte, mapIdPos *DBPosIndex, pathToFile string) error {
	f, err := os.OpenFile(pathToFile,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	pos, ok := (*mapIdPos)[id]
	log.Println("Should delete line starting at pos:", pos)
	if !ok {
		return errors.New("the message id doesn't exist in the index map")
	}

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
