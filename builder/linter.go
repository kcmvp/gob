package builder

import (
	"bufio"
	"bytes"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/samber/lo"

	"github.com/jedib0t/go-pretty/v6/table"

	"github.com/fatih/color"
	"github.com/kcmvp/gob/boot"
	"github.com/thedevsaddam/gojsonq/v2"
)

const (
	IssueNode        = "Issues"
	lintCfg          = ".golangci.yml"
	lintCmd          = "golangci-lint"
	lintModule       = "github.com/golangci/golangci-lint/cmd/golangci-lint"
	LintHTMLReport   = "lint.html"
	LintJSONReport   = "lint.json"
	LintOutputReport = "lint.out"
)

var _ boot.Installer = (*Linter)(nil)

//go:embed template/.golangci.yml
var golangCiTmp string

//go:embed template/golang_lint.tmpl
var reportTpl string

type Linter struct {
	boot.Installer
}

func newLinter() *Linter {
	return &Linter{
		boot.NewInstallable(lintModule, lintCmd, lintCfg, linterVersion),
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
func (linter *Linter) scan(session *boot.Session, builder *Builder, command boot.Command) error {
	ver := builder.Config().GetString(fmt.Sprintf("%s.%s", boot.CfgPrefix, linter.CfgVerKey()))
	if len(ver) < 1 {
		return errors.New("lint is not setup")
	}
	ver, err := linter.Install(ver)
	if err != nil {
		return fmt.Errorf("failed to install golangci-linter %s: %w", ver, err)
	}
	os.Chdir(builder.RootDir())
	args := []string{"run", "-v", "--out-format", "json", "./..."}
	changedOnly := (builder.Initializer() != boot.None || !session.GetFlagBool(command, "all")) && command != boot.Report
	if changedOnly {
		args = append(args, "--new-from-rev", "HEAD~")
	} else {
		// if '--fix' is set in the command line then keep it, otherwise it should be always false
		args = append(args, "--fix", "false")
	}
	vCmd := fmt.Sprintf("%s-%s", linter.Cmd(), linter.Format(ver))
	log.Printf("Scan with %s-%s", linter.Cmd(), ver)
	session.SaveCtxValue(command, strings.Join(append(args, vCmd), " "))
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

	sc := bufio.NewScanner(stderr)
	output, err := os.Create(filepath.Join(builder.TargetDir(), session.Specified(LintOutputReport)))
	defer output.Close()
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

	// write data to json
	jsonData, _ := io.ReadAll(stdout)
	var prettyJSON bytes.Buffer
	err = json.Indent(&prettyJSON, jsonData, "", "\t")
	if err != nil {
		return fmt.Errorf("failed to save lint json report: %w", err)
	}
	jsonReport := filepath.Join(builder.TargetDir(), session.Specified(LintJSONReport))
	jf, _ := os.Create(jsonReport)
	defer jf.Close()
	jf.Write(prettyJSON.Bytes())

	// html report
	jq := gojsonq.New().FromString(prettyJSON.String()).From(IssueNode)
	issues := jq.Count()
	if issues > 0 {
		jq = jq.Select("FromLinter as Linter", "Text as Message", "SourceLines as Code", "Pos.Filename as File", "Pos.Line as Line", "Pos.Column as Column")
		data := jq.Get()
		//@todo need to be removed
		/*
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

			report := filepath.Join(builder.TargetDir(), session.Specified(LintHTMLReport))
			html, err := os.Create(report)
			defer html.Close()
			if err != nil {
				return fmt.Errorf("failed to create lint report: %w", err)
			}
			err = t.Execute(html, data)
			if err != nil {
				return fmt.Errorf("failed to generate lint report: %w", err)
			}
		*/
		log.Printf("lint report is generated at %s\n", filepath.Join(builder.TargetDir(), LintHTMLReport))
		msg := fmt.Sprintf("Total %d linter issues are found", issues)
		log.Println(color.RedString(msg))
		tableReport(builder.TargetDir(), data)
		if changedOnly {
			return errors.New(msg)
		}
	} else {
		log.Println(color.GreenString("no linter issues are found"))
	}
	return nil
}

func splitIntoGroup(msg string, size int) []string {
	sub := ""
	var subs []string
	runes := bytes.Runes([]byte(msg))
	l := len(runes)
	for i, r := range runes {
		sub += string(r)
		if (i+1)%size == 0 {
			subs = append(subs, sub)
			sub = ""
		} else if (i + 1) == l {
			subs = append(subs, sub)
		}
	}
	return subs
}

func tableReport(dir string, data interface{}) error {
	ct := table.Table{}
	ct.SetTitle("Lint Issues Report")
	style := table.StyleDefault
	style.Options.DrawBorder = true
	style.Options.SeparateRows = true
	style.Options.SeparateColumns = true
	style.HTML.CSSClass = table.DefaultHTMLCSSClass
	ct.SetStyle(style)
	consoleRows := lo.Map(data.([]interface{}), func(t interface{}, i int) table.Row {
		tm := t.(map[string]interface{})
		filePath := tm["File"].(string)
		if len(filePath) > 30 {
			strings.Split(filePath, string(os.PathSeparator))
			filePath = strings.Join(strings.Split(filePath, string(os.PathSeparator)), fmt.Sprintf("%s%s", string(os.PathSeparator), "\n"))
		}
		code := tm["Code"].([]interface{})[0].(string)
		code = strings.TrimSpace(code)
		msg := tm["Message"].(string)
		msg = strings.TrimSpace(msg)
		return table.Row{i + 1, filePath, strings.TrimSpace(fmt.Sprintf("%v:%v", tm["Line"], tm["Column"])), tm["Linter"], strings.Join(lo.Slice(splitIntoGroup(code, 40), 0, 2), "\n"), strings.Join(splitIntoGroup(msg, 70), "\n")}
	})
	ct.AppendHeader(table.Row{"#", "File", "Line", "Linter", "Code", "Message"})
	ct.AppendRows(consoleRows)
	fmt.Println(ct.Render())
	tableRows := lo.Map(data.([]interface{}), func(t interface{}, i int) table.Row {
		tm := t.(map[string]interface{})
		filePath := tm["File"].(string)
		code := tm["Code"].([]interface{})[0].(string)
		code = strings.TrimSpace(code)
		msg := tm["Message"].(string)
		msg = strings.TrimSpace(msg)
		return table.Row{i + 1, filePath, strings.TrimSpace(fmt.Sprintf("%v:%v", tm["Line"], tm["Column"])), tm["Linter"], code, msg}
	})
	ct.ResetRows()
	ct.AppendRows(tableRows)
	html := ct.RenderHTML()
	htmlReport, err := os.Create(filepath.Join(dir, LintHTMLReport))
	if err != nil {
		return err //nolint
	}
	_, err = htmlReport.WriteString(html)
	if err != nil {
		return err //nolint
	}
	err = htmlReport.Close()
	return err //nolint
}
