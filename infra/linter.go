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
	msg := "lint fullScan source code"
	args := []string{"run", "-v", "./...", "--out-format=json"}
	if !fullScan {
		args = append(args, "--new-from-rev=HEAD")
		msg = "lint source for changes"
	}
	log.Println(msg)
	vCmd := fmt.Sprintf("%s-%s", linter.Cmd(), ver)
	if output, err := exec.Command(vCmd, args...).CombinedOutput(); err != nil { //nolint:nestif
		// save the report
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
			log.Fatalln(color.RedString("runs into error %s", err.Error()))
		}
		if err = os.WriteFile(file, prettyJSON.Bytes(), os.ModePerm); err != nil {
			log.Fatalln(color.RedString("failed to save lint report %s", err.Error()))
		}
		jq := gojsonq.New().FromString(prettyJSON.String()).From(IssueNode)
		issues := jq.Count()
		msg = fmt.Sprintf("%d of issues are found, you can get detail report at %s", issues, file)
		if failOnIssue {
			log.Fatalln(color.RedString(msg))
		} else {
			log.Println(msg)
		}
	} else {
		log.Println(color.GreenString("no linter issues are found"))
	}
}

func VerifyLinter(dir string) {
	f := filepath.Join(dir, linter.output)
	jq := gojsonq.New().File(f).From(IssueNode)
	issues := jq.Count()

	jq = gojsonq.New().File(f).From(IssueNode)
	files := jq.Distinct("Pos.Filename").Count()
	if issues > 0 {
		log.Fatalln(color.RedString("%d of new issues are found in %d files, please refer to %s", issues, files, f))
	} else {
		log.Println(color.GreenString("no issues are found"))
	}
}

func GenerateLintCfg(data interface{}, trunk bool) {
	GenerateFile(golangCiTmp, lintCfg, data, trunk)
}
