package linter

import (
	"bufio"
	_ "embed"
	"fmt"
	"github.com/fatih/color"
	"github.com/kcmvp/gbt/infra"
	"log"
	"os"
	"os/exec"
	"strings"
)

const (
	IssueNode   = "Issues"
	Cfg         = ".golangci.yml"
	cmd         = "golangci-lint"
	changedOnly = "--new-from-rev=HEAD"
	module      = "github.com/golangci/golangci-lint/cmd/golangci-lint"
)

var linter infra.Installable

func init() {
	linter = infra.NewInstallable(module, cmd, func(cmd string) string {
		ver := ""
		output, err := exec.Command(cmd, "version").CombinedOutput()
		if err == nil {
			ver = strings.Fields(string(output))[3]
		} else {
			log.Fatalln(color.RedString("%s, %+v", string(output), err))
		}
		return ver
	})
}

func Install(ver string) (string, error) {
	return linter.Install(ver)
}

func ConfiguredVer() (string, error) {
	var ver string
	var err error
	f, err := os.Open(Cfg)
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
			msg := color.RedString("missed version in %s", Cfg)
			log.Println(msg)
			err = fmt.Errorf(msg)
		}
	}
	return ver, err
}

func Scan(dir string, changesFile, formatOnly bool) {
	if ver, err := ConfiguredVer(); err == nil {
		if ver, err = linter.Install(ver); err != nil {
			args := []string{"run", "-v", "./...", dir, "--out-format=json", "--new-from-rev=HEAD"}
			vCmd := fmt.Sprintf("%s-%s", linter.Cmd(), ver)
			output, _ := exec.Command(vCmd, args...).CombinedOutput()
			if !formatOnly {
				fmt.Printf(string(output))
			}
		} else {
			log.Fatalln(color.RedString("can't find %s, please run 'gbt setup linter' to setup linter", Cfg))
		}
	}
}

/*
   func (linter *Linter) parse(project *builder.Project, data []byte) {
   	file := filepath.Join(project.TargetDir(), linter.command+".json")
   	sc := bufio.NewScanner(strings.NewReader(string(data)))
   	var line string
   	// don't print detail in the commit message hook
   	for sc.Scan() {
   		line = sc.Text()
   		if strings.HasPrefix(line, "{\"Issues\"") {
   			break
   		} else if project.GitHook() != hook.CommitMessage {
   			cline := line
   			if strings.HasPrefix(line, "level=warning") {
   				cline = color.YellowString(line)
   			}
   			log.Println(cline)
   		}
   	}

   	var prettyJSON bytes.Buffer
   	if err := json.Indent(&prettyJSON, []byte(line), "", "\t"); err != nil {
   		log.Fatalln(color.RedString("runs into error %s", err))
   	}
   	os.WriteFile(file, prettyJSON.Bytes(), os.ModePerm)

   	jq := gojsonq.New().FromString(prettyJSON.String()).From(IssueNode)

   	issue := project.Report().LinterIssues

   	issue.Issues = jq.Count()
   	obj := jq.GroupBy("FromLinter").Get()
   	if m, ok := obj.(map[string][]interface{}); ok {
   		for k, v := range m {
   			issue.Detail[k] = len(v)
   		}
   	}

   	jq = gojsonq.New().FromString(prettyJSON.String()).From(IssueNode)
   	issue.Files = jq.Distinct("Pos.Filename").Count()
   	if issue.Issues > 0 { //nolint:nestif
   		log.Println(color.YellowString("total %d issues are found in %d files", issue.Issues, issue.Files))
   		if project.GitHook() == hook.CommitMessage {
   			jq = gojsonq.New().FromString(prettyJSON.String()).From(IssueNode).Select("FromLinter", "Text", "Pos.Filename as File", "Pos.Line as Line", "Pos.Column as Column")
   			lines := jq.Get()
   			if v, ok := lines.([]interface{}); ok {
   				for _, m := range v {
   					if mm, ok := m.(map[string]interface{}); ok {
   						log.Println(color.RedString("[%v]: %v %v:%v - %v", mm["FromLinter"], mm["File"], mm["Line"], mm["Column"], mm["Text"]))
   					}
   				}
   			}
   		}
   		log.Println(color.YellowString("please check %s for detail", filepath.Join("target", "golangci-lint.json")))
   	} else {
   		log.Println(color.GreenString("no new issues are found"))
   	}
   }

   func (linter *Linter) Setup() {
   	if _, err := os.Stat(".golangci.yml"); err != nil {
   		log.Fatalln(color.RedString("missed %s", linterCfg))
   	}
   }
*/
