////go:build gbt

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
)

const MsgPattern = "#[0-9]{1,7}:.*"
const Separator = "#[0-9]{1,7}:"
const MinLength = 10

func main() {
	input, _ := os.ReadFile(os.Args[1])
	checkMessage(string(input))
	scan()
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

func scan() {
	if _, file, _, ok := runtime.Caller(1); ok {
		dir := filepath.Dir(file)
		builder := filepath.Join(dir, "builder.go")
		if _, err := os.Stat(builder); err == nil {
			out, err := exec.Command("go", "run", builder, "-event=message_hook").CombinedOutput()
			fmt.Println(string(out))
			if err != nil {
				os.Exit(1)
			} else {
				os.Exit(0)
			}
		} else {
			fmt.Printf("%s does not exist\n", builder)
		}
	} else {
		fmt.Println("runs into error")
	}
	os.Exit(1)
}
