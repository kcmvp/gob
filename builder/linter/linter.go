package linter

import (
	_ "embed"
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

func Scan(formatOnly bool) {
	//linter.validate()
	//if err := os.MkdirAll(p.TargetDir(), os.ModePerm); err != nil {
	//	fmt.Printf("failed to create directory %v", err)
	//	os.Exit(1)
	//}
	//args := append(linter.args, fmt.Sprintf("%s/...", p.ModuleDir()))
	//if p.gitHook.event == hook.CommitMessage {
	//	args = append(linter.args, scanChangedFlag)
	//}
	//output, _ := exec.Command(linter.command, args...).CombinedOutput()
	//linter.parse(p, output)
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

func (linter *Linter) validate() {
	if _, err := os.Stat(".golangci.yml"); err != nil {
		log.Fatalln(color.RedString("missed %s", linterCfg))
	}
}
*/
