package main

import (
	"bufio"
	"encoding/csv"
	"io"
	"os"
	"strings"
	"time"
)

type DBPosIndex map[[16]byte]int64

type DBChronFinder struct {
	ChronIndex 	*DBChronIndex
	TimeArr   	*DBTimeArr
}

type DBChronIndex = map[int64][16]byte // we need an alias instead of type definition (see appendUniques)
type DBTimeArr = []int64


func fillPositionIndex(pathToFile string) (*DBPosIndex, error) {
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
	for i:=0; !eof; i++ {
		line, err = r.ReadString('\n')
		if err != io.EOF {
			if err != nil {
				return nil, err
			}
		} else {
			eof = true
		}
		record := strings.Split(line, ",")
		mapIdPos[idToHex16byte(record[0])] = posBytes
		posBytes += int64(len(line))
	}
	return &mapIdPos, nil
}

// fillChronIndArr fills up the creation time array and the chronological index
func fillChronIndArr(pathToFile string) (DBChronFinder, error) {
	mapTimeId := make(DBChronIndex)
	timeArr := make(DBTimeArr,0)

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
		check(err)
		mapTimeId[t.UnixNano()] = idToHex16byte(record[0])
		timeArr = append(timeArr, t.UnixNano())
	}
	return DBChronFinder{&mapTimeId, &timeArr}, nil
}

