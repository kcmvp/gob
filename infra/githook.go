package infra

import (
	"embed"
	"encoding/json"
	"fmt"
	"github.com/go-git/go-git/v5"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/fatih/color"
)

var hooks = []string{"pre-commit", "commit-msg", "pre-push"}

//go:embed template/*.tmpl
var templateDir embed.FS

func Hooks() map[string]string {
	m := map[string]string{}
	for _, hook := range hooks {
		m[hook] = strings.Replace(hook, "-", "_", 1)
	}
	return m
}

type Hook struct {
	Target string
	Type   string
}

func SetupHook(root, scriptDir string, genNew bool) error {
	var err error
	var tf []byte
	gitDir := filepath.Join(root, ".git")
	for s, g := range Hooks() {
		gof := fmt.Sprintf("%s.go", g)
		abs, _ := filepath.Abs(filepath.Join(scriptDir, gof))
		if _, err = os.Stat(abs); err != nil {
			if !genNew {
				continue
			}
			if tf, err = templateDir.ReadFile(filepath.Join("template", fmt.Sprintf("%s.tmpl", g))); err == nil {
				GenerateFile(string(tf), abs, nil, false)
			}
		} else if tf, err = templateDir.ReadFile(filepath.Join("template", "hook.tmpl")); err == nil {
			GenerateFile(string(tf), filepath.Join(gitDir, "hooks", s), Hook{abs, s}, true)
		}
	}
	return err
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

func PrePush(version, target string, repo *git.Repository) {
	input, _ := os.ReadFile(os.Args[1])
	rep := regexp.MustCompile(`\r?\n`)
	commitMsg := rep.ReplaceAllString(string(input), "")
	reg, err := regexp.Compile(version)
	if err == nil && !reg.MatchString(commitMsg) {
		log.Fatalln(color.RedString("commit message must follow %s", version))
	}
	// check the consistent between version and target
	s := Report{}
	t := Report{}
	data, err := os.ReadFile(version)
	if err != nil {
		log.Fatalln(color.RedString("can not open file %s", version))
	}
	err = json.Unmarshal(data, &s)
	if err != nil {
		log.Fatalln(color.RedString("incorrect data format %s", version))
	}

	data, err = os.ReadFile(target)
	if err != nil {
		log.Fatalln(color.RedString("can not open file %s", target))
	}
	err = json.Unmarshal(data, &t)
	if err != nil {
		log.Fatalln(color.RedString("incorrect data format %s", target))
	}

	for k, v := range s.Packages {
		if t.Packages[k] != v {
			log.Fatalln(color.RedString("value of %s is not the same between %s and %s", k, s, t))
		}
	}

	if s.Tests != t.Tests {
		log.Fatalln(color.RedString("number of the test is not the same between %s and %s", s, t))
	}
	// check the degrade

}
