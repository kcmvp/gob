package builder

import (
	"bufio"
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/thedevsaddam/gojsonq/v2"
)

type (
	ArgF       func() []string
	ParseF     func(issue *Issue, data []byte, file string)
	Validation func() error
)

type Dependency struct {
	command    string
	args       ArgF
	parser     ParseF
	validation Validation
}

var golangCiLinter = &Dependency{
	command: "golangci-lint",
	args: func() []string {
		args := []string{"run", "-v", "./...", "--out-format=json"}
		return args
	},
	parser:     golangCiParser,
	validation: golangciValidation,
}

func (s *Dependency) Scan(p *Project, args ...string) {
	s.validation()
	if err := os.MkdirAll(p.TargetDir(), os.ModePerm); err != nil {
		fmt.Printf("failed to create directory %v", err)
		os.Exit(1)
	}
	dir := filepath.Join(p.TargetDir(), s.command+".json")
	args = append(s.args(), args...)
	args = append(args, fmt.Sprintf("%s/...", p.ModuleDir()))
	msg, _ := exec.Command(s.command, args...).CombinedOutput()
	s.parser(&p.quality.Issues, msg, dir)
}

var golangCiParser ParseF = func(issue *Issue, data []byte, file string) {
	sc := bufio.NewScanner(strings.NewReader(string(data)))
	var line string
	for sc.Scan() {
		line = sc.Text()
		if strings.HasPrefix(line, "{\"Issues\"") {
			break
		} else {
			cline := line
			if strings.HasPrefix(line, "level=warning") {
				cline = color.YellowString(line)
			}
			log.Println(cline)
		}
	}
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, []byte(line), "", "\t"); err != nil {
		if e, ok := err.(*json.SyntaxError); ok {
			log.Printf("syntax error at byte offset %d", e.Offset)
		}
	}
	os.WriteFile(file, prettyJSON.Bytes(), os.ModePerm)

	jq := gojsonq.New().FromString(string(prettyJSON.Bytes())).From("Issues")
	issue.Issues = jq.Count()
	obj := jq.GroupBy("FromLinter").Get()
	if m, ok := obj.(map[string][]interface{}); ok {
		for k, v := range m {
			issue.Linters[k] = len(v)
		}
	}
	jq = gojsonq.New().FromString(prettyJSON.String()).From("Issues")
	issue.Files = jq.Distinct("Pos.Filename").Count()
	if issue.Issues > 0 {
		log.Println(color.YellowString("total %d issues are found in %d files", issue.Issues, issue.Files))
		log.Println(color.YellowString("please check %s for detail", filepath.Join("target", "golangci-lint.json")))
	} else {
		log.Println(color.CyanString("no new issues are found"))
	}
}

var golangciValidation = func() error {
	_, err := os.Stat(".golangci.yml")
	return err
}
