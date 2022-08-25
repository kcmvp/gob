package infra

import (
	"bufio"
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/fatih/color"
	"github.com/thedevsaddam/gojsonq/v2"
)

const (
	IssueNode  = "Issues"
	lintCfg    = ".golangci.yml"
	lintCmd    = "golangci-lint"
	lintModule = "github.com/golangci/golangci-lint/cmd/golangci-lint"
)

var _ Installable = (*golangCiLinter)(nil)

//go:embed template/.golangci.yml
var golangCiTmp string

//go:embed template/golang-lint.tmpl
var reportTpl string

var linter golangCiLinter

type golangCiLinter struct {
	Installable
	output string
}

var linterVersion = func(cmd string) string {
	ver := ""
	output, err := exec.Command(cmd, "version").CombinedOutput()
	if err == nil {
		ver = strings.Fields(string(output))[3]
	} else {
		log.Fatalln(color.RedString("%s, %s", string(output), err.Error()))
	}
	return ver
}

func init() {
	ins := NewInstallable(lintModule, lintCmd, linterVersion)
	linter = golangCiLinter{
		ins,
		fmt.Sprintf("%s.html", lintCmd),
	}
}

func Install(ver string) (string, error) {
	return linter.Install(ver)
}

func ConfiguredLinterVer() (string, error) {
	var ver string
	var err error
	f, err := os.Open(lintCfg)
	defer f.Close()
	if err == nil {
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := scanner.Text()
			if items := strings.Split(line, "version:"); len(items) == 2 {
				ver = strings.TrimSpace(items[1])
				log.Printf("linter is configured as %s \n", ver)
				break
			}
		}
		if ver == "" {
			msg := color.RedString("missed version in %s", lintCfg)
			log.Println(msg)
			err = fmt.Errorf(msg)
		}
	}
	return ver, err
}

//nolint:cyclop
func LintScan(targetDir string, fullScan bool, failOnIssue bool) {
	ver, err := ConfiguredLinterVer()
	if err != nil {
		log.Fatalln(color.RedString("lint scan: %s", err.Error()))
	}
	ver, err = linter.Install(ver)
	if err != nil {
		log.Fatalln(color.RedString("failed to install golangci-linter %s: %s", ver, err.Error()))
	}
	msg := "lint all source code"
	args := []string{"run", "-v", "./...", "--out-format=json"}
	if !fullScan {
		args = append(args, "--new-from-rev=HEAD")
		msg = "lint changed source code"
	}
	log.Println(msg)
	vCmd := fmt.Sprintf("%s-%s", linter.Cmd(), ver)
	output, err := exec.Command(vCmd, args...).CombinedOutput()
	if err == nil {
		log.Println("no linter issues are found")
		return
	}
	// save the LintReport

	file := filepath.Join(targetDir, linter.output)
	sc := bufio.NewScanner(strings.NewReader(string(output)))
	r := regexp.MustCompile(`".*"`)
	n := regexp.MustCompile(`config_reader|lintersdb|before processing:.*after processing:`)
	var line string
	for sc.Scan() {
		line = sc.Text()
		if strings.HasPrefix(line, "{\"Issues\"") {
			break
		} else if fullScan && (strings.HasPrefix(line, "level=warning") || n.MatchString(line)) {
			msg = strings.ReplaceAll(r.FindString(line), "\"", "")
			if strings.HasPrefix(line, "level=warning") {
				msg = color.YellowString(msg)
			}
			log.Println(msg)
		}
	}

	var prettyJSON bytes.Buffer
	if err = json.Indent(&prettyJSON, []byte(line), "", "\t"); err != nil {
		log.Fatalln(color.RedString("runs into error: %s", err.Error()))
	}
	jq := gojsonq.New().FromString(prettyJSON.String()).From(IssueNode)
	jq = jq.Select("FromLinter as Linter", "Text as Message", "SourceLines as Code", "Pos.Filename as File", "Pos.Line as Line", "Pos.Column as Column")
	data := jq.Get()
	funcMap := template.FuncMap{
		"add": func(i int) int {
			return i + 1
		},
		"concat": func(s []interface{}) string {
			var s1 []string
			for _, i := range s {
				s1 = append(s1, fmt.Sprintf("%v", i))
			}
			return strings.TrimSpace(strings.Join(s1, "\n"))
		},
	}
	t, err := template.New("report").Funcs(funcMap).Parse(reportTpl)
	checkError(err)
	f, err := os.Create(file)
	checkError(err)
	err = t.Execute(f, data)
	checkError(err)
	msg = fmt.Sprintf("lint report is generated at %s", file)
	if failOnIssue {
		log.Fatalln(color.RedString(msg))
	} else {
		log.Println(msg)
	}
}

func GenerateLintCfg(data interface{}, trunk bool) {
	if err := GenerateFile(golangCiTmp, lintCfg, data, trunk); err != nil {
		log.Fatalln(color.RedString("failed to generate lint config:%s", err.Error()))
	}
}

func checkError(err error) {
	if err != nil {
		log.Fatalln(color.RedString("runs into error: %s", err.Error()))
	}
}
