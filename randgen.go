package main

import (
	"encoding/hex"
	"fmt"
	"math/rand"
	"strings"
)

func randomIdStr16(randNow *rand.Rand) (string, error) {
	bytes := make([]byte, 16)
	if _, err := randNow.Read(bytes); err != nil {
		return "", err
	}
	hexStrB := []byte(hex.EncodeToString(bytes))
	str := fmt.Sprint(
		strings.ToUpper(string(hexStrB[0:8])),"-",
		strings.ToUpper(string(hexStrB[8:12])),"-",
		strings.ToUpper(string(hexStrB[12:16])),"-",
		strings.ToUpper(string(hexStrB[16:20])),"-",
		strings.ToUpper(string(hexStrB[20:32])))
	return str, nil
}
