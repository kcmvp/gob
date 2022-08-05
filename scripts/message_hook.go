//go:build gbt

package main

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/kcmvp/gbt/builder"
	"log"
	"os"
	"regexp"
	"strings"
)

const MsgPattern = "#[0-9]{1,7}:.*"
const Separator = "#[0-9]{1,7}:"
const MinLength = 10

func main() {
	input, _ := os.ReadFile(os.Args[1])
	checkMessage(string(input))
	project := builder.NewProject().Clean().Scan()
	if project.Quality().LiterIssues.Files > 0 {
		log.Print(color.RedString("failed to commit the code"))
		os.Exit(1)
	}
	os.Exit(0)
}

func checkMessage(msg string) {
	reg, err := regexp.Compile(MsgPattern)
	sp, _ := regexp.Compile(Separator)
	if err != nil {
		fmt.Println(fmt.Sprintf("internal error %v", err))
		os.Exit(1)
	}
	if !reg.MatchString(msg) {
		log.Print(color.RedString("commit message must follow format #{number}: xxxxxx"))
		os.Exit(1)
	}
	items := sp.Split(msg, -1)
	// check message length
	if len(strings.TrimSpace(items[1])) < MinLength {
		log.Println(color.RedString("commit message is at least %d characters\n", MinLength))
		os.Exit(1)
	}
}
