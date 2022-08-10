package builder

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/fatih/color"
)

const (
	lineCoverageReport   = "line.data"
	methodCoverageReport = "method.data"
	rawTestReport        = "test.data"
	quality              = "quality.json"
	scriptDir            = "scripts"
	scriptLine           = "go run %s $1 $2\n"
	buildTarget          = "target"
)

type Quality struct {
	Methods      int
	Tests        int
	Coverage     Coverage
	LinterIssues *LinterIssue
}

type Coverage struct {
	Method string
	Line   string
}

// @todo rename to Linter
// @ add Linter version.
type LinterIssue struct {
	Files  int
	Issues int
	Detail map[string]int
}

type Project struct {
	ctx       context.Context
	moduleDir string
	// rootDir    string
	scriptsDir string
	targetDir  string
	quality    *Quality
	// hook            HookEvent
	gitHook *GitHook
}

type testCase struct {
	Package string
	Test    string
	Action  string
	Output  string
	Elapsed float64
}

func moduleDir() string {
	_, file, _, ok := runtime.Caller(2)
	if ok {
		p := filepath.Dir(file)
		for {
			if _, err := os.ReadFile(filepath.Join(p, "go.mod")); err == nil {
				return p
			} else {
				p = filepath.Dir(p)
			}
		}
	}
	panic("Can't figure out module directory")
}

func NewProject(cfg *HookCfg) *Project {
	project := &Project{
		moduleDir: moduleDir(),
		quality: &Quality{
			LinterIssues: &LinterIssue{
				Detail: map[string]int{},
			},
		},
	}
	project.targetDir = filepath.Join(project.moduleDir, buildTarget)
	project.scriptsDir = filepath.Join(project.moduleDir, scriptDir)
	err := os.MkdirAll(project.targetDir, os.ModePerm)
	FatalIfError(err)
	err = os.MkdirAll(project.scriptsDir, os.ModePerm)
	FatalIfError(err)
	// project.rootDir = projectRoot(project.moduleDir)
	project.setupHook(cfg)
	return project
}

func (project *Project) ModuleDir() string {
	return project.moduleDir
}

func (project *Project) TargetDir() string {
	return project.targetDir
}

func (project *Project) Quality() *Quality {
	return project.quality
}

func (project *Project) WithCtx(ctx context.Context) *Project {
	project.ctx = ctx
	return project
}

func (project *Project) Ctx() context.Context {
	if project.ctx == nil {
		project.ctx = context.Background()
	}
	return project.ctx
}

func (project *Project) GitHook() HookEvent {
	return project.gitHook.event
}

func (project *Project) setupHook(cfg *HookCfg) {
	// validate folders
	gitHook := newGitHook(project.moduleDir, cfg)
	project.gitHook = gitHook

	pcs := make([]uintptr, 10)
	n := runtime.Callers(0, pcs)
	pcs = pcs[:n]
	frames := runtime.CallersFrames(pcs)
	for {
		frame, more := frames.Next()
		if !more {
			break
		}
		for k, v := range HookMap {
			if strings.HasSuffix(frame.File, v) {
				gitHook.event = k
				break
			}
		}
	}
	gitHook.validate()
}

func (project *Project) buildTestReport() {
	file, err := os.Open(filepath.Join(project.targetDir, rawTestReport))
	if err != nil {
		log.Fatalln(color.RedString("failed to open the file %v \n", filepath.Join(project.targetDir, rawTestReport)))
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)

	testSet := map[string]bool{}
	for scanner.Scan() {
		text := scanner.Text()
		c := testCase{}
		json.Unmarshal([]byte(text), &c)
		testSet[c.Test] = true
	}
	project.quality.Tests = len(testSet)

	mc, err := os.Open(filepath.Join(project.targetDir, methodCoverageReport))
	if err != nil {
		log.Fatalln(color.RedString("failed to open the file %v", filepath.Join(project.targetDir, methodCoverageReport)))
	}
	defer mc.Close()
	testedMethod := 0
	scanner = bufio.NewScanner(mc)
	for scanner.Scan() {
		text := scanner.Text()
		items := strings.Fields(text)
		coverage, _ := strconv.ParseFloat(strings.TrimRight(items[2], "%"), 64)
		//coverage /= 100
		if strings.EqualFold(items[0], "total:") {
			project.quality.Coverage.Line = items[2]
		} else {
			project.quality.Methods++
			if coverage > 0 {
				testedMethod++
			}
		}
	}
	//project.quality.Coverage.Method = math.Floor(float64(testedMethod)/float64(project.quality.Methods)*1000) / 1000
	project.quality.Coverage.Method = fmt.Sprintf("%.2f", float64(testedMethod)*100/float64(project.quality.Methods))
}

func (project *Project) Clean() *Project {
	log.Println("clean build directory ......")
	if err := os.RemoveAll(project.targetDir); err != nil {
		log.Fatalln(color.RedString("failed to delete %s\n", project.targetDir))
	}
	return project
}

// Test run the test with -race, -cover, -fuzz and -bench.
func (project *Project) Test(args ...string) *Project {
	log.Println("run unit test ......")
	os.Chdir(project.moduleDir)
	os.MkdirAll(project.targetDir, os.ModePerm)
	params := []string{"test", "-v", "-json", "-coverprofile", filepath.Join(project.targetDir, lineCoverageReport), "./..."}
	if len(args) > 0 {
		params = append(params, args...)
	}
	out, err := exec.Command("go", params...).CombinedOutput()
	if err != nil {
		sc := bufio.NewScanner(strings.NewReader(string(out)))
		for sc.Scan() {
			line := sc.Text()
			if !strings.HasPrefix(line, "{\"Time\":") {
				log.Println(color.RedString(line))
			}
		}
		os.Exit(1)
	}

	os.WriteFile(filepath.Join(project.targetDir, rawTestReport), out, os.ModePerm)
	//  go tool cover -func ./targetDir/coverage.data
	params = []string{"tool", "cover", "-func", filepath.Join(project.targetDir, lineCoverageReport)}
	out, _ = exec.Command("go", params...).CombinedOutput()
	os.WriteFile(filepath.Join(project.targetDir, methodCoverageReport), out, os.ModePerm)
	project.buildTestReport()
	log.Println(color.CyanString("total tests :%d, line coverage: %f, method coverage %f", project.quality.Tests, project.quality.Coverage.Line, project.quality.Coverage.Method))
	return project
}

// Build walk from module directory and run build command for each executable
// and place the executable at ${project_root}/bin; in case there are more than one executable.
func (project *Project) Build(files ...string) *Project {
	targetFiles := files
	if len(targetFiles) == 0 {
		targetFiles = append(targetFiles, "main.go")
	}
	log.Println("build project ......")
	os.MkdirAll(project.targetDir, os.ModePerm)
	filepath.Walk(project.moduleDir, func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		for _, t := range targetFiles {
			if strings.EqualFold(info.Name(), t) {
				if output, err := exec.Command("go", "build", "-o", project.targetDir, path).CombinedOutput(); err != nil {
					fmt.Println(string(output))
				}
			}
		}
		return nil
	})
	return project
}

func (project *Project) Scan(args ...string) *Project {
	log.Println("scan source code ......")
	project.gitHook.beforeScan(args...)
	linter.Scan(project)

	project.saveReport(filepath.Join(project.targetDir, quality))
	project.gitHook.afterScan(project, args...)

	return project
}

func (project *Project) saveReport(file string) {
	data, err := json.Marshal(project.quality)
	FatalIfError(err)
	var prettyJSON bytes.Buffer
	err = json.Indent(&prettyJSON, data, "", "\t")
	FatalIfError(err)
	err = os.WriteFile(file, prettyJSON.Bytes(), os.ModePerm)
	FatalIfError(err)
}

// func projectRoot(dir string) string {
//	tmp := dir
//	for tmp != string(os.PathSeparator) {
//		if _, err := os.Stat(filepath.Join(tmp, git.GitDirName)); err == nil {
//			return tmp
//		} else {
//			tmp = filepath.Dir(tmp)
//		}
//	}
//	if tmp == string(os.PathSeparator) {
//		log.Println(color.YellowString("%s is not valid git project", dir))
//		dir = ""
//	}
//	return dir
//}

func FatalIfError(err error) {
	if err == nil {
		return
	}
	log.Println(color.RedString("runs into error %+v", err))
	pcs := make([]uintptr, 10)
	n := runtime.Callers(0, pcs)
	pcs = pcs[:n]
	frames := runtime.CallersFrames(pcs)
	frame, more := frames.Next()
	for more {
		log.Println(color.RedString("%s#%d", frame.File, frame.Line))
		frame, more = frames.Next()
	}
	os.Exit(1)
}
