package infra

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

//var GolangLint golangCiLinter
var linter golangCiLinter

type golangCiLinter struct {
	Installable
	targetDir string
	output    string
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
		"",
		fmt.Sprintf("%s.json", lintCmd),
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

func LintScan(target string, all bool) {
	linter.targetDir = target
	if ver, err := ConfiguredLinterVer(); err == nil {
		if ver, err = linter.Install(ver); err == nil {
			msg := "lint all source code"
			args := []string{"run", "-v", "./...", "--out-format=json"}
			if !all {
				args = append(args, "--new-from-rev=HEAD")
				msg = "lint source for changes"
			}
			log.Println(msg)
			vCmd := fmt.Sprintf("%s-%s", linter.Cmd(), ver)
			output, _ := exec.Command(vCmd, args...).CombinedOutput()
			// save the report
			file := filepath.Join(linter.targetDir, linter.output)
			sc := bufio.NewScanner(strings.NewReader(string(output)))
			r, _ := regexp.Compile(`".*"`)
			n, _ := regexp.Compile(`config_reader|lintersdb|before processing:.*after processing:`)
			var line string
			for sc.Scan() {
				line = sc.Text()
				if strings.HasPrefix(line, "{\"Issues\"") {
					break
				} else {
					if strings.HasPrefix(line, "level=warning") || n.MatchString(line) {
						msg = strings.ReplaceAll(r.FindString(line), "\"", "")
						if strings.HasPrefix(line, "level=warning") {
							msg = color.YellowString(msg)
						}
						log.Println(msg)
					}
				}
			}

			var prettyJSON bytes.Buffer
			if err = json.Indent(&prettyJSON, []byte(line), "", "\t"); err != nil {
				log.Fatalln(color.RedString("runs into error %s", err))
			}
			os.WriteFile(file, prettyJSON.Bytes(), os.ModePerm)
			log.Printf("lint report is generated at %s \n", file)
		} else {
			log.Fatalln(color.RedString("can't find %s, please run 'gbt setup linter' to setup linter", lintCfg))
		}
	}
}

func VerifyLinter(halt bool) {
	f := filepath.Join(linter.targetDir, linter.output)
	jq := gojsonq.New().File(f).From(IssueNode)
	issues := jq.Count()

	jq = gojsonq.New().File(f).From(IssueNode)
	files := jq.Distinct("Pos.Filename").Count()
	if issues > 0 {
		msg := fmt.Sprintf("%d of new issues are found in %d files, please refer to %s", issues, files, f)
		if halt {
			log.Fatalln(color.RedString(msg))
		} else {
			log.Println(color.YellowString(msg))
		}
	} else {
		log.Println(color.GreenString("no issues are found"))
	}
}

func GenerateLintCfg(data interface{}, trunk bool) {
	GenerateFile(golangCiTmp, lintCfg, data, trunk)
}
