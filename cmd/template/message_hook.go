//go:build gbt

package main

import (
	"fmt"
	"github.com/kcmvp/gbt/builder"
	"os"
	"regexp"
)

const MsgPattern = "#[0-9]{1,7}:.*"
const Separator = "#[0-9]{1,7}:"
const MinLength = 10

func main() {
	input, _ := os.ReadFile(os.Args[1])
	checkMessage(string(input))
	project := builder.NewProject().Scan()
	if project.Quality().Issues.Files > 0 {
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
		fmt.Println("commit message must follow format #{number}: xxxxxx")
		os.Exit(1)
	}
	items := sp.Split(msg, -1)
	// check message length
	if len(items[1]) < MinLength {
		fmt.Println(fmt.Sprintf("commit message is at least %d characters", MinLength))
		os.Exit(1)
	}
}
