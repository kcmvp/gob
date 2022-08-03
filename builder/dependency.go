package builder

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/thedevsaddam/gojsonq/v2"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type (
	ArgF   func() []string
	ParseF func(issue *Issue, data []byte, file string)
)

var golangCiLinter = &Dependency{
	module:  "github.com/golangci/golangci-lint/cmd/golangci-lint@latest",
	command: "golangci-lint",
	args: func() []string {
		args := []string{"run", "-v", "./...", "--out-format=json"}
		fmt.Printf("%s %s \n", "golangci-lint", args)
		return args
	},
	parser: golangCiParser,
}

type Dependency struct {
	module  string
	command string
	args    ArgF
	parser  ParseF
}

func (s *Dependency) Exec(p *Project) {
	dir := filepath.Join(p.TargetDir(), s.command+".json")
	args := append(s.args(), fmt.Sprintf("%s/...", p.ModuleDir()))
	msg, _ := exec.Command(s.command, args...).CombinedOutput()
	fmt.Printf(string(msg))
	s.parser(&p.quality.Issues, msg, dir)
}

var golangCiParser ParseF = func(issue *Issue, data []byte, file string) {
	items := strings.Split(strings.Trim(string(data), "\n"), "\n")
	item := items[0]
	for _, item = range items {
		if strings.HasPrefix(item, "{\"Issues\"") {
			break
		}
	}
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, []byte(item), "", "\t"); err != nil {
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

	jq = gojsonq.New().FromString(string(prettyJSON.Bytes())).From("Issues")
	issue.Files = jq.Distinct("Pos.Filename").Count()
}
