package githook

import (
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/fatih/color"
)

var hooks = []string{"pre-commit", "commit-msg", "pre-push"}

func Hooks() map[string]string {
	m := map[string]string{}
	for _, hook := range hooks {
		m[hook] = strings.Replace(hook, "-", "_", 1)
	}
	return m
}

func ValidateCommitMsg(pattern string) {
	input, _ := os.ReadFile(os.Args[1])
	rep := regexp.MustCompile(`\r?\n`)
	commitMsg := rep.ReplaceAllString(string(input), "")
	reg, err := regexp.Compile(pattern)
	if err == nil && !reg.MatchString(commitMsg) {
		log.Fatalln(color.RedString("commit message must follow %s", pattern))
	}
}

func ValidateCoverage(pattern string) {
	input, _ := os.ReadFile(os.Args[1])
	rep := regexp.MustCompile(`\r?\n`)
	commitMsg := rep.ReplaceAllString(string(input), "")
	reg, err := regexp.Compile(pattern)
	if err == nil && !reg.MatchString(commitMsg) {
		log.Fatalln(color.RedString("commit message must follow %s", pattern))
	}
	//@todo
	// 1: verify the json is the same scripts/coverage.json & target/coverage.json
	// 2: verify the coverage does not degress

}
