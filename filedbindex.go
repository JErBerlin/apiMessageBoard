// filedbindex.go provides indexing objects and functions for a flat-file database with records of type Message.String()
package main

import (
	"bufio"
	"encoding/csv"
	"io"
	"os"
	"strings"
	"time"
)

// a position index is a map from 16 digit hex id to a int64 position (bytes positions from beginning of file)
type DBPosIndex map[[16]byte]int64

// a chronological finder contains a chronological index and a time array
type DBChronFinder struct {
	ChronIndex *DBChronIndex
	TimeArr    *DBTimeArr
}

// a chronological index is a map from int64 time (nanoseconds) to 16 digit hex id
type DBChronIndex = map[int64][16]byte // we need an alias instead of type definition (see appendUniques)

// a time array contains int64 time (nanoseconds)
type DBTimeArr = []int64


// FillPositionIndex fills a position index for the db, as map from 16 hex id to int64 position
func FillPositionIndex(pathToFile string) (*DBPosIndex, error) {
	mapIdPos := make(DBPosIndex)

	f, err := os.Open(pathToFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := bufio.NewReader(f)
	line, err := r.ReadString('\n')
	if err != nil {
		return nil, err
	}
	lenHeader := len(line)
	posBytes := int64(lenHeader)

	eof := false
	for i := 0; !eof; i++ {
		line, err = r.ReadString('\n')
		if err != io.EOF {
			if err != nil {
				return nil, err
			}
		} else {
			eof = true
		}
		record := strings.Split(line, ",")
		id,_ := IdToHex16byte(record[0])
		mapIdPos[id] = posBytes
		posBytes += int64(len(line))
	}
	return &mapIdPos, nil
}

// FillChronIndArr fills up the creation time array and the chronological index
// the chronological index is a map from int64 time (nanoseconds) to 16 digit hex id
func FillChronIndArr(pathToFile string) (DBChronFinder, error) {
	mapTimeId := make(DBChronIndex)
	timeArr := make(DBTimeArr, 0)

	csvFile, err := os.Open(pathToFile)
	if err != nil {
		return DBChronFinder{}, err
	}

	defer csvFile.Close()

	r := csv.NewReader(csvFile)
	// first line is header
	record, err := r.Read()
	if err != nil {
		return DBChronFinder{}, err
	}
	for {
		record, err = r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return DBChronFinder{}, err
		}

		form := timeFormat
		t, err := time.Parse(form, record[4])
		if err != nil {
			return DBChronFinder{}, err
		}
		mapTimeId[t.UnixNano()], _ = IdToHex16byte(record[0])
		timeArr = append(timeArr, t.UnixNano())
	}
	return DBChronFinder{&mapTimeId, &timeArr}, nil
}
