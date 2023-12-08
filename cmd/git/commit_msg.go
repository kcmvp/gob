//go:build gob

package main

import (
	"os"
	"regexp"
)

// must start with '#' and then follows numbers(ticket number) and ':' then the message body
const commitMsgPattern = `^#[0-9]+:\s*.{10,}$`

func main() {
	input, _ := os.ReadFile(os.Args[1])
	regex := regexp.MustCompile(`\r?\n`)
	commitMsg := regex.ReplaceAllString(string(input), "")
	regex = regexp.MustCompile(commitMsgPattern)
	if !regex.MatchString(commitMsg) {
		os.Exit(1)
	}
	os.Exit(0)
}
