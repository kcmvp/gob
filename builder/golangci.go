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
	ParseF func(issue *LiterIssue, data []byte, file string)
)

const (
	IssueNode   = "Issues"
	golangCiCfg = ".golangci.yml"
)

type GolangCi struct {
	command string
	args    []string
	parser  ParseF
}

var golangCiLinter = &GolangCi{
	command: "golangci-lint",
	args:    []string{"run", "-v", "./...", "--out-format=json"},
}

func (s *GolangCi) Scan(p *Project, args ...string) {
	s.validate()
	if err := os.MkdirAll(p.TargetDir(), os.ModePerm); err != nil {
		fmt.Printf("failed to create directory %v", err)
		os.Exit(1)
	}
	dir := filepath.Join(p.TargetDir(), s.command+".json")
	args = append(s.args, args...)
	args = append(args, fmt.Sprintf("%s/...", p.ModuleDir()))
	msg, _ := exec.Command(s.command, args...).CombinedOutput()
	fmt.Println(p.quality.LiterIssue)
	fmt.Println(dir, msg)
	//s.parser(&p.quality.LiterIssue, msg, dir)
}

var golangCiParser ParseF = func(issue *LiterIssue, data []byte, file string) {
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

	jq := gojsonq.New().FromString(string(prettyJSON.Bytes())).From(IssueNode)
	issue.Issues = jq.Count()
	obj := jq.GroupBy("FromLinter").Get()
	if m, ok := obj.(map[string][]interface{}); ok {
		for k, v := range m {
			issue.Linters[k] = len(v)
		}
	}
	jq = gojsonq.New().FromString(prettyJSON.String()).From(IssueNode)
	issue.Files = jq.Distinct("Pos.Filename").Count()
	if issue.Issues > 0 {
		log.Println(color.YellowString("total %d issues are found in %d files", issue.Issues, issue.Files))
		log.Println(color.YellowString("please check %s for detail", filepath.Join("target", "golangci-lint.json")))
	} else {
		log.Println(color.GreenString("no new issues are found"))
	}
}

func (s *GolangCi) validate() {
	if _, err := os.Stat(".golangci.yml"); err != nil {
		log.Fatalln(color.RedString("missed %s", golangCiCfg))
	}
}
