//go:build gob

package main

import (
	"github.com/fatih/color"
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
		color.Red("Error: commit message must follow %s", commitMsgPattern)
		os.Exit(1)
	}
	os.Exit(0)
}
