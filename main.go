package main

import (
	"log"
)

const (
	PathToMessagesFile = "./messages.csv"
	timeFormat = "2006-01-02T15:04:05-07:00"
)

func main() {

	startRouter()
}

// TODO: implement specific error handling in functions, decide if some fatal err case
func check(err error) {
	if err!= nil {
		log.Println(err)
	}
}