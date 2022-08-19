package githook

import (
	"github.com/fatih/color"
	"log"
	"os"
	"regexp"
	"strings"
)

var hooks = []string{"pre-commit", "commit-msg", "pre-push"}

func Hooks() map[string]string {
	m := map[string]string{}
	for _, hook := range hooks {
		m[hook] = strings.Replace(hook, "-", "_", 1)
	}
	return m
}

func CommitMsg(pattern string) {
	input, _ := os.ReadFile(os.Args[1])
	rep := regexp.MustCompile(`\r?\n`)
	commitMsg := rep.ReplaceAllString(string(input), "")
	reg, err := regexp.Compile(pattern)
	if err == nil && !reg.MatchString(commitMsg) {
		log.Fatalln(color.RedString("commit message must follow %s", pattern))
	}
}

func PrePush(pattern string) {
	input, _ := os.ReadFile(os.Args[1])
	rep := regexp.MustCompile(`\r?\n`)
	commitMsg := rep.ReplaceAllString(string(input), "")
	reg, err := regexp.Compile(pattern)
	if err == nil && !reg.MatchString(commitMsg) {
		log.Fatalln(color.RedString("commit message must follow %s", pattern))
	}
}
