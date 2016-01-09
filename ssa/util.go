package ssa

import (
	"bytes"
	"fmt"
	"strings"
)

const (
	escapeValues  = "\n\t\r\f\b\"\\"
	escapeLetters = "ntrfb\"\\"
)

func EscapeString(str string) string {
	sanBuf := new(bytes.Buffer)

	unsanStr := []byte(str)

	for _, char := range unsanStr {
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') || strings.ContainsRune(" _!@$%^&*(){}[]|:;<>?,./~`", rune(char)) {
			sanBuf.WriteByte(char)
		} else if index := strings.IndexRune(escapeValues, rune(char)); index >= 0 {
			sanBuf.WriteRune('\\')
			sanBuf.WriteByte(escapeLetters[index])
		} else {
			sanBuf.WriteString(fmt.Sprintf("\\%03o", char))
		}
	}

	return sanBuf.String()
}
