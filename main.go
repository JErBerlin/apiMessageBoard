package main

import (
	"log"
)

const (
	PathToMessagesFile = "./messages.csv"
	timeFormat = "2006-01-02T15:04:05-07:00"
	testLen = 150000
	testWriteLen = 150000
)

func main() {

	startServing()

}

// TODO: implement specific error handling in functions, decide if some fatal err case
func check(err error) {
	if err!= nil {
		log.Println(err)
	}
}