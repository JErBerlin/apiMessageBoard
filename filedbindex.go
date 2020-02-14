package main

import (
	"bufio"
	"encoding/csv"
	"io"
	"os"
	"strings"
	"time"
)

type DBIndex map[[16]byte]int64

type DBChronIndex map[int64][16]byte

func fillPositionIndex(pathToFile string) *DBIndex {
	mapIdPos := make(DBIndex)

	f, err := os.Open(pathToFile)
	check(err)
	defer f.Close()

	r := bufio.NewReader(f)
	line, err := r.ReadString('\n')
	check(err)
	lenHeader := len(line)
	posBytes := int64(lenHeader)

	eof := false
	for i:=0; i < testLen && !eof; i++ {
		line, err = r.ReadString('\n')
		if err != io.EOF {
			check(err)
		} else {
			eof = true
		}
		record := strings.Split(line, ",")
		mapIdPos[idToHex16byte(record[0])] = posBytes
		posBytes += int64(len(line))
	}
	return &mapIdPos
}

// fillChronIndArr fills up the creation time array and the chronological index
func fillChronIndArr(pathToFile string) (*DBChronIndex, *[]int64) {
	mapTimeId := make(DBChronIndex)
	timeArr := make([]int64,0)

	csvFile, err := os.Open(pathToFile)
	check(err)
	defer csvFile.Close()

	r := csv.NewReader(csvFile)
	// first line is header
	record, err := r.Read()
	check(err)
	for i:=0; i<testLen; i++{
		record, err = r.Read()
		if err == io.EOF {
			break
		}
		check(err)

		form := timeFormat
		t, err := time.Parse(form, record[4])
		check(err)
		mapTimeId[t.UnixNano()] = idToHex16byte(record[0])
		timeArr = append(timeArr, t.UnixNano())
	}
	return &mapTimeId, &timeArr
}

