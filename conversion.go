package main

import (
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
)

func idToHex16byte(str string) [16]byte {
	reg, err := regexp.Compile("[^a-zA-Z0-9]+")
	check(err)
	idHex, err := hex.DecodeString(reg.ReplaceAllString(str, ""))
	check(err)

	var idHex16 [16]byte
	copy(idHex16[:], idHex)
	return idHex16
}

func idHex16toStr(idHex16 [16]byte) string {
	idHex := make([]byte, 16)
	copy(idHex, idHex16[:])
	hexStrB := []byte(hex.EncodeToString(idHex))
	return fmt.Sprint(
		strings.ToUpper(string(hexStrB[0:8])), "-",
		strings.ToUpper(string(hexStrB[8:12])), "-",
		strings.ToUpper(string(hexStrB[12:16])), "-",
		strings.ToUpper(string(hexStrB[16:20])), "-",
		strings.ToUpper(string(hexStrB[20:32])))
}
