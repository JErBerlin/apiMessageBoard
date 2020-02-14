package main

import (
	"fmt"
	"reflect"
	"strings"
	"time"
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

// String returns the values Message in a format suitable for a record in a csv file
func (m Message) String() string {
	timeStr := m.CreationTime.Format("2006-01-02T15:04:05-07:00")
	return fmt.Sprintf("%s,%s,%s,\"%s\",%s", m.Id,m.Name,m.Email,m.Text,timeStr)
}
