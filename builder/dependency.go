package builder

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/thedevsaddam/gojsonq/v2"
)

const (
	colorRed   = "\033[31m"
	colorReset = "\033[0m"
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

func (s *Dependency) Exec(p *Project, args ...string) {
	s.validation()
	dir := filepath.Join(p.TargetDir(), s.command+".json")
	args = append(s.args(), args...)
	args = append(args, fmt.Sprintf("%s/...", p.ModuleDir()))
	msg, _ := exec.Command(s.command, args...).CombinedOutput()
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

	jq = gojsonq.New().FromString(prettyJSON.String()).From("Issues")
	issue.Files = jq.Distinct("Pos.Filename").Count()

	jq = gojsonq.New().FromString(prettyJSON.String()).From("Issues")
	v := jq.Select("FromLinter", "Pos.Filename as File", "Pos.Line as Line", "Pos.Column as Column", "SourceLines as Code", "Text as Msg")
	if items, ok := v.Get().([]interface{}); ok {
		for _, i := range items {
			if o, ok := i.(map[string]interface{}); ok {
				fmt.Printf(colorRed+"%s#%v:%v [%s]:%s\n", o["File"], o["Line"], o["Column"], o["FromLinter"], o["Msg"])
				// if c, ok := o["Code"].([]interface{}); ok {
				//	str := fmt.Sprintf("%v", c[0])
				//	fmt.Printf(colorReset+"%v\n", strings.TrimSpace(str))
				//}
			}
		}
	}
	fmt.Println(colorReset + "")
}

var golangciValidation = func() error {
	_, err := os.Stat(".golangci.yml")
	return err
}
