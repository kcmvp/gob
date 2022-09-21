package scaffolds

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

	"golang.org/x/net/html"

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
func (linter *Linter) scan(session *boot.Session, builder *Project, command boot.Command) error {
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
		log.Printf("lint report is generated at %s\n", filepath.Join(builder.TargetDir(), LintHTMLReport))
		msg := fmt.Sprintf("Total %d linter issues are found", issues)
		log.Println(color.RedString(msg))
		saveLintReport(session, builder.TargetDir(), data, changedOnly)
		if changedOnly {
			return errors.New(msg)
		}
	} else {
		log.Println(color.GreenString("No linter issues are found"))
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

func saveLintReport(session *boot.Session, dir string, data interface{}, consoleOutput bool) error {
	ct := table.Table{}
	ct.SetTitle("Lint Issues Report")
	ct.AppendHeader(table.Row{"#", "File", "Line", "Linter", "Code", "Message"})
	style := table.StyleDefault
	style.Options.DrawBorder = true
	style.Options.SeparateRows = true
	style.Options.SeparateColumns = true
	style.HTML.CSSClass = table.DefaultHTMLCSSClass
	ct.SetStyle(style)
	var err error
	if consoleOutput {
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
		ct.AppendRows(consoleRows)
		fmt.Println(ct.Render())
	} else {
		tableRows := lo.Map(data.([]interface{}), func(t interface{}, i int) table.Row {
			tm := t.(map[string]interface{})
			filePath := tm["File"].(string)
			code := tm["Code"].([]interface{})[0].(string)
			code = strings.TrimSpace(code)
			msg := tm["Message"].(string)
			msg = strings.TrimSpace(msg)
			return table.Row{i + 1, filePath, strings.TrimSpace(fmt.Sprintf("%v:%v", tm["Line"], tm["Column"])), tm["Linter"], code, msg}
		})
		ct.AppendRows(tableRows)
		content := ct.RenderHTML()
		var htmlReport *os.File
		htmlReport, err = os.Create(filepath.Join(dir, session.Specified(LintHTMLReport)))
		if err != nil {
			return err //nolint
		}
		tableStyle := ` <style>
		  table, th, td {
		    border: 1px solid;
		  }
		 </style>`
		root, _ := html.Parse(strings.NewReader(tableStyle + content))
		if err = html.Render(htmlReport, root); err != nil {
			log.Println(color.RedString("Failed to save lint html report:%^", err.Error()))
		}
		err = htmlReport.Close()
	}
	return err //nolint
}
