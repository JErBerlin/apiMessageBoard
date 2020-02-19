// message.go defines the basic the structure message and provides methods to stringify its fields
package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"reflect"
	"strings"
	"time"
)

type Message struct {
	Id           string
	Name         string
	Email        string
	Text         string
	CreationTime time.Time
}

// PrintFields return field names as they appear in the type definition
func (m Message) PrintFields() string {

	var fields []string
	for i := 0; i < reflect.TypeOf(m).NumField(); i++ {
		fields = append(fields,
			strings.ToLower(reflect.TypeOf(m).Field(i).Name))
	}
	return strings.Join(fields, " ")
}

// String returns the values Message in a format suitable for a record in a csv file
func (m Message) String() string {
	timeStr := m.CreationTime.Format(timeFormat)
	return fmt.Sprintf("%s,%s,%s,\"%s\",%s", m.Id, m.Name, m.Email, m.Text, timeStr)
}

// NewFromRecord returns a new Message with the information provided by a record from the csv file
func NewFromRecord (record []string) (Message, error) {
	t, err := time.Parse(timeFormat, record[4])
	if err != nil {
		return Message{}, err
	}

	msg := Message{
		Id:           record[0],
		Name:         record[1],
		Email:        record[2],
		Text:         record[3],
		CreationTime: t,
	}

	return msg, nil
}

// NewFromJSON returns a new Message with the information provided by a JSON object
// if the json msg doesn't have a time, the time field it is set to present time
// if the json msg doesn't have id, the id is randomly generated with a helper function
func NewFromJSON (msgJSON []byte) (Message, error) {
	msg := Message{}
	json.Unmarshal(msgJSON, &msg)

	newTime := time.Now()
	newSource := rand.NewSource(int64(newTime.Nanosecond()))
	randNow := rand.New(newSource)

	if msg.CreationTime.IsZero() {
		msg.CreationTime = newTime
	}
	if msg.Id == "" {
		newIdStr, err := RandomIdStr16(randNow)
		if err != nil {
			return Message{}, err
		}
		msg.Id = newIdStr
	}
	return msg, nil
}
