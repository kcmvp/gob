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

const (
	IssueNode       = "Issues"
	linterCfg       = ".golangci.yml"
	linterCommand   = "golangci-lint"
	scanChangedFlag = "--new-from-rev=HEAD"
)

type Linter struct {
	command string
	args    []string
}

var linter = &Linter{
	command: linterCommand,
	args:    []string{"run", "-v", "./...", "--out-format=json"},
}

func (linter *Linter) Scan(p *Project) {
	linter.validate()
	if err := os.MkdirAll(p.TargetDir(), os.ModePerm); err != nil {
		fmt.Printf("failed to create directory %v", err)
		os.Exit(1)
	}
	args := linter.args
	if p.scanChanged {
		args = append(linter.args, scanChangedFlag)
	}
	args = append(args, fmt.Sprintf("%s/...", p.ModuleDir()))
	output, _ := exec.Command(linter.command, args...).CombinedOutput()
	linter.parse(p, output)
}

func (linter *Linter) parse(project *Project, data []byte) {
	file := filepath.Join(project.TargetDir(), linter.command+".json")
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

	issue := project.Quality().LinterIssues

	issue.Issues = jq.Count()
	obj := jq.GroupBy("FromLinter").Get()
	if m, ok := obj.(map[string][]interface{}); ok {
		for k, v := range m {
			issue.Detail[k] = len(v)
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

func (linter *Linter) validate() {
	if _, err := os.Stat(".golangci.yml"); err != nil {
		log.Fatalln(color.RedString("missed %s", linterCfg))
	}
}
