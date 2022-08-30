package infra

import (
	"bufio"
	_ "embed"
	"errors"
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

var _ Installable = (*Linter)(nil)

//go:embed template/.golangci.yml
var golangCiTmp string

//go:embed template/golang_lint.tmpl
var reportTpl string

var linter = &Linter{
	NewInstallable(lintModule, lintCmd, linterVersion),
	fmt.Sprintf("%s.html", lintCmd),
	fmt.Sprintf("%s.out", lintCmd),
}

type Linter struct {
	Installable
	report string
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

func InstallLinter(ver string) (string, error) {
	return linter.Install(ver)
}

func GetLinterVer(root string) (string, error) {
	var ver string
	var err error
	f, err := os.Open(filepath.Join(root, lintCfg))
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

func LintScan(cfgDir, targetDir string, fullScan bool, failOnIssue bool) error {
	ver, err := GetLinterVer(cfgDir)
	if err != nil {
		return fmt.Errorf("lint scan: %w", err)
	}
	ver, err = linter.Install(ver)
	if err != nil {
		return fmt.Errorf("failed to install golangci-linter %s: %w", ver, err)
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
	if err != nil {
		return fmt.Errorf("failed to get error pipe from linter command: %w", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get error pipe from linter command: %w", err)
	}
	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to execute linter command: %w", err)
	}
	report := filepath.Join(targetDir, linter.report)
	sc := bufio.NewScanner(stderr)
	output, err := os.OpenFile(filepath.Join(targetDir, linter.output), os.O_APPEND|os.O_CREATE|os.O_WRONLY, os.ModePerm) //nolint
	defer output.Close()                                                                                                  //nolint                                                                                    //nolint
	if err != nil {
		return fmt.Errorf("failed to create linter output file: %w", err)
	}
	for sc.Scan() {
		tmpLine := sc.Text()
		_, err = output.WriteString(tmpLine + "\n")
		if err != nil {
			return fmt.Errorf("failed to create linter output: %w", err)
		}
		log.Println(tmpLine)
	}
	if err != nil {
		return fmt.Errorf("failed to get lint standard out: %w", err)
	}
	jq := gojsonq.New().Reader(stdout).From(IssueNode)
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
	if err != nil {
		return fmt.Errorf("failed to parse lint report template: %w", err)
	}
	f, err := os.Create(report)
	defer f.Close() //nolint
	if err != nil {
		return fmt.Errorf("failed to create lint report: %w", err)
	}
	err = t.Execute(f, data)
	if err != nil {
		return fmt.Errorf("failed to generate lint report: %w", err)
	}
	log.Printf("lint report is generated at %s", report)
	if issues > 0 {
		msg = fmt.Sprintf("total %d linter issues are found", issues)
		if failOnIssue {
			return errors.New(msg)
		} else {
			log.Println(color.YellowString(msg))
		}
	} else {
		log.Println(color.GreenString("no linter issues are found"))
	}
	return nil
}

func GenLinterCfg(data interface{}, trunk bool) {
	if err := GenerateFile(golangCiTmp, lintCfg, data, trunk); err != nil {
		log.Fatalln(color.RedString("failed to generate lint config:%s", err.Error()))
	}
}
