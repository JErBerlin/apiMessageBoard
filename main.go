package main

import (
	"fmt"
	"log"
	"math/rand"
	"sort"
	"time"
)

const (
	PathToMessagesFile = "./messages.csv"
	timeFormat = "2006-01-02T15:04:05-07:00"
	testLen = 3000000
	testWriteLen = 300000
)

func main() {
	// do position and date indexing, as preparation for preloading of messages
	log.Println("Indexing at start..")
	mapIdPos := fillPositionIndex(PathToMessagesFile)
	_ = mapIdPos
	mapTimeId, timeArr := fillChronIndArr(PathToMessagesFile)
	_ = mapTimeId
	_ = timeArr

	// write messages
	oneMessage := readMessageFromFileById(idToHex16byte("080B78DA-262D-EA54-391F-71FE92109F09"), mapIdPos, PathToMessagesFile)
	timeNow := time.Now()
	randInt64 := int64(timeNow.Nanosecond())
	newSource := rand.NewSource(randInt64)
	randNow := rand.New(newSource)
	for i:=0; i < testWriteLen; i++ {
		log.Print("iter:",i,",")
		newTime := time.Now()
		rand.Seed(int64(i))
		afterXseconds := newTime.Add(time.Second*time.Duration(rand.Int63n(testWriteLen)))
		newIdStr, _ := randomIdStr16(randNow)
		oneMessage.Id = newIdStr
		oneMessage.CreationTime = afterXseconds
		err := writeMessageToFile(oneMessage, PathToMessagesFile)
		check(err)
	}

	/*
	// read all messages
	messages := readMessagesFromFile(PathToMessagesFile)
	_ = messages
	*/

	// do indexing again
	log.Println("Re-indexing after write operation")
	mapIdPos = fillPositionIndex(PathToMessagesFile)
	_ = mapIdPos
	mapTimeId, timeArr = fillChronIndArr(PathToMessagesFile)
	_ = mapTimeId

	// sort the times anti-chronologically
	sort.Slice(*timeArr, func(i, j int) bool{ return (*timeArr)[i] > (*timeArr)[j]})

	fmt.Println(len(*timeArr), "messages:")
	for i:=0; i< testLen && i < len(*timeArr); i++ {
		t := (*timeArr)[i]
		id := (*mapTimeId)[t]

		oneMessage := readMessageFromFileById(id, mapIdPos, PathToMessagesFile)
		fmt.Printf("%4d: \t%v -- %s\n", i+1, oneMessage.CreationTime.Format("02/01/2006- 15:04:05"), oneMessage.Id)
		// fmt.Println(t, "--",idHex16toStr((*mapTimeId)[oneMessage.CreationTime.UnixNano()]))
	}

	// startServingMessages(messages)

}

func check(e error) {
	if e != nil {
		log.Fatal(e)
	}
}