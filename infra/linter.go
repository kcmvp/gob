package infra

import (
	"bufio"
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"os"
	"os/exec"
	"path/filepath"
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

//go:embed template/golang_lint.tmpl
var reportTpl string

var linter golangCiLinter

type golangCiLinter struct {
	Installable
	report string
	output string
	root   string
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

func SetupLinterService(ctx context.Context) {
	if dir, err := root(ctx); err == nil {
		ins := NewInstallable(lintModule, lintCmd, linterVersion)
		linter = golangCiLinter{
			ins,
			fmt.Sprintf("%s.html", lintCmd),
			fmt.Sprintf("%s.out", lintCmd),
			dir,
		}
	} else {
		log.Println(color.YellowString("%s", err.Error()))
	}
}

func Install(ver string) (string, error) {
	return linter.Install(ver)
}

func ConfiguredLinterVer() (string, error) {
	var ver string
	var err error
	f, err := os.Open(filepath.Join(linter.root, lintCfg))
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

	cmd := exec.Command(vCmd, args...)
	stderr, err := cmd.StderrPipe()
	CheckError(err, "Failed to get error pipe from linter command")
	stdout, err := cmd.StdoutPipe()
	CheckError(err, "Failed to get output pipe from linter command")
	err = cmd.Start()
	CheckError(err, "Failed to execute linter command")
	report := filepath.Join(targetDir, linter.report)
	sc := bufio.NewScanner(stderr)
	output, err := os.OpenFile(filepath.Join(targetDir, linter.output), os.O_APPEND|os.O_CREATE|os.O_WRONLY, os.ModePerm) //nolint
	CheckError(err, "Failed to create linter output file")
	defer output.Close() //nolint                                                                                    //nolint
	for sc.Scan() {
		tmpLine := sc.Text()
		_, err = output.WriteString(tmpLine + "\n")
		CheckError(err, "Failed to create linter output")
		log.Println(tmpLine)
	}

	CheckError(err, "Failed to get lint standard out")
	sc = bufio.NewScanner(stdout)
	var line string
	for sc.Scan() {
		line = sc.Text()
	}

	var prettyJSON bytes.Buffer
	err = json.Indent(&prettyJSON, []byte(line), "", "\t")
	CheckError(err, "Failed to indent lint report")
	jq := gojsonq.New().FromString(prettyJSON.String()).From(IssueNode)
	issues := jq.Count()
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
	CheckError(err, "Failed to parse lint report template")
	f, err := os.Create(report)
	defer f.Close() //nolint
	CheckError(err, "Failed to create lint report")
	err = t.Execute(f, data)
	CheckError(err, "Failed to generate lint report")
	log.Printf("lint report is generated at %s", report)
	if issues > 0 {
		msg = fmt.Sprintf("total %d linter issues are found", issues)
		if failOnIssue {
			log.Fatalln(color.RedString(msg))
		} else {
			log.Println(color.YellowString(msg))
		}
	} else {
		log.Println(color.GreenString("no linter issues are found"))
	}
}

func GenerateLintCfg(data interface{}, trunk bool) {
	if err := GenerateFile(golangCiTmp, lintCfg, data, trunk); err != nil {
		log.Fatalln(color.RedString("failed to generate lint config:%s", err.Error()))
	}
}
