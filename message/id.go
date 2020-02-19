// id.go provides helper functions for the field id of the type message
package message

import (
	"encoding/hex"
	"fmt"
	"math/rand"
	"regexp"
	"strings"
)

// IdToHex16byte turns a 16 digit hex string id into an array [16]bytes
func IdToHex16byte(str string) ([16]byte, error) {
	reg, err := regexp.Compile("[^a-zA-Z0-9]+")
	if err != nil {
		return [16]byte{}, nil
	}
	idHex, err := hex.DecodeString(reg.ReplaceAllString(str, ""))
	if err != nil {
		return [16]byte{}, nil
	}

	var idHex16 [16]byte
	copy(idHex16[:], idHex)
	return idHex16, nil
}

// IdHex16toStr turns a [16]byte id into a string (with separators)
func IdHex16toStr(idHex16 [16]byte) string {
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

// RandomIdStr16 generates a 16 digits hex id in form of string (with separators)
func RandomIdStr16(randNow *rand.Rand) (string, error) {
	bytes := make([]byte, 16)
	if _, err := randNow.Read(bytes); err != nil {
		return "", err
	}
	hexStrB := []byte(hex.EncodeToString(bytes))
	str := fmt.Sprint(
		strings.ToUpper(string(hexStrB[0:8])), "-",
		strings.ToUpper(string(hexStrB[8:12])), "-",
		strings.ToUpper(string(hexStrB[12:16])), "-",
		strings.ToUpper(string(hexStrB[16:20])), "-",
		strings.ToUpper(string(hexStrB[20:32])))
	return str, nil
}