package builder

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
	"time"

	"github.com/fatih/color"
	"github.com/kcmvp/gob/boot"
	"github.com/thedevsaddam/gojsonq/v2"
)

const (
	IssueNode  = "Issues"
	lintCfg    = ".golangci.yml"
	lintCmd    = "golangci-lint"
	lintModule = "github.com/golangci/golangci-lint/cmd/golangci-lint"
)

var _ boot.Installer = (*Linter)(nil)

//go:embed template/.golangci.yml
var golangCiTmp string

//go:embed template/golang_lint.tmpl
var reportTpl string

type Linter struct {
	boot.Installer
	report string
	output string
}

func newLinter() *Linter {
	return &Linter{
		boot.NewInstallable(lintModule, lintCmd, lintCfg, linterVersion),
		fmt.Sprintf("%s.html", lintCmd),
		fmt.Sprintf("%s.out", lintCmd),
	}
}

var linterVersion = func(name string) (string, string) {
	ver := ""
	cmd := exec.Command(name, "version")
	output, err := cmd.CombinedOutput()
	if err == nil {
		ver = strings.Fields(string(output))[3]
	} else {
		log.Println(color.RedString("%s, %s", string(output), err.Error()))
	}
	return ver, cmd.Path
}

// nolint
func (linter *Linter) scan(builder *Builder, command boot.Command) error {
	ver := builder.Config().GetString(linter.CfgVerKey())
	if len(ver) < 1 {
		return errors.New("lint is not setup")
	}
	ver, err := linter.Install(ver)
	if err != nil {
		return fmt.Errorf("failed to install golangci-linter %s: %w", ver, err)
	}
	os.Chdir(builder.RootDir())
	args := []string{"run", "-v", "--out-format", "json", "./..."}
	changeOnly := builder.Initializer() != boot.None || !boot.GetFlag[bool](command, "all")
	flags := boot.AllFlags(command)
	log.Println(flags)
	if changeOnly {
		args = append(args, "--new-from-rev", "HEAD~")
	}
	vCmd := fmt.Sprintf("%s-%s", linter.Cmd(), linter.Format(ver))
	log.Printf("Scan with %s-%s", linter.Cmd(), ver)
	boot.SaveExecCtx(command, strings.Join(append(args, vCmd), " "))
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

	suffix := fmt.Sprintf(".%d.tmp", time.Now().UnixMilli())
	defer rename(builder.TargetDir(), suffix)
	sc := bufio.NewScanner(stderr)
	//output, err := os.OpenFile(filepath.Join(builder.TargetDir(), linter.output, suffix), os.O_APPEND|os.O_CREATE|os.O_WRONLY, os.ModePerm) //nolint
	output, err := os.Create(filepath.Join(builder.TargetDir(), fmt.Sprintf("%s%s", linter.output, suffix)))
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
	err = output.Close()
	if err != nil {
		return fmt.Errorf("failed to create linter output file: %w", err)
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

	report := filepath.Join(builder.TargetDir(), fmt.Sprintf("%s%s", linter.report, suffix))
	f, err := os.Create(report)
	if err != nil {
		return fmt.Errorf("failed to create lint report: %w", err)
	}
	err = t.Execute(f, data)
	if err != nil {
		return fmt.Errorf("failed to generate lint report: %w", err)
	}
	log.Printf("lint report is generated at %s\n", report)
	if issues > 0 {
		msg := fmt.Sprintf("total %d linter issues are found", issues)
		if changeOnly {
			return errors.New(msg)
		} else {
			log.Println(color.YellowString(msg))
		}
	} else {
		log.Println(color.GreenString("no linter issues are found"))
	}
	f.Close()
	return nil
}
