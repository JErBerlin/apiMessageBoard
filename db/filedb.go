// filedb.go provides writing and reading functions for records of type Message in a flat-file database
package db

import (
	"encoding/csv"
	"errors"
	"fmt"
	"github.com/JErBerlin/back_message_board/message"
	"io"
	"os"
)

// Record represents a line in a csv file
type Record []string

// WriteMessageToFile write a message to the end of the file at pathToFile
func WriteMessageToFile(msg message.Message, pathToFile string) error {
	f, err := os.OpenFile(pathToFile,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.WriteString("\n" + fmt.Sprint(msg)); err != nil {
		return err
	}
	return nil
}


// ReadMessageFromFileById returns the record identified by id
func ReadMessageFromFileById(id [16]byte, mapIdPos *DBPosIndex, pathToFile string) (Record, error) {
	f, err := os.Open(pathToFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	pos, ok := (*mapIdPos)[id]
	if !ok {
		return nil, errors.New("the message id doesn't exist in the index")
	}
	_, err = f.Seek(pos, 0) // offset = pos, whence = 0 (reference is origin of the file)
	if err != nil {
		return nil, err
	}

	r := csv.NewReader(f)
	record, err := r.Read()
	if err != nil {
		if err == io.EOF {
			return nil, errors.New("EOF reading message from file")
		}
		return nil, err
	}
	return record, nil
}

// ReplaceMessageInFileById leaves the old message identified by id unchanged, and writes another one with the same id
// and cloned fields, but with a new text, at the end of the file
func ReplaceMessageInFileById(msg message.Message, id [16]byte, mapIdPos *DBPosIndex, pathToFile string) error {
	f, err := os.OpenFile(pathToFile,
		os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	record, err := ReadMessageFromFileById(id, mapIdPos, pathToFile)
	if err != nil {
		return err
	}

	oldMessage, err := message.NewFromRecord(record)
	if err != nil {
		return err
	}

	// copy relevant fields (name, from old message to new
	msg.Name = oldMessage.Name
	msg.Email = oldMessage.Email
	msg.CreationTime = oldMessage.CreationTime

	err = WriteMessageToFile(msg, pathToFile)
	if err != nil {
		return errors.New("replace operation of the in-file message failed: could not write")
	}
	return nil
}


