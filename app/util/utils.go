package util

import (
	"strings"
)

// CleanStr Function to remove non-printable characters
func CleanStr(str string) string {
	cleanStr := func(r rune) rune {
		if r >= 32 && r != 127 {
			return r
		}
		return -1
	}
	return strings.Map(cleanStr, str)
}
