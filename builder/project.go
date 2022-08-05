package builder

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"math"
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
	messageHook          = "message_hook.go"
	pushHook             = "push_hook.go"
)

type Quality struct {
	Methods      int
	Tests        int
	Coverage     Coverage
	LinterIssues *LinterIssue
}

type Coverage struct {
	Method float64
	Line   float64
}

// @todo rename to Linter
// @ add Linter version.
type LinterIssue struct {
	Files  int
	Issues int
	Detail map[string]int
}

type Project struct {
	ctx             context.Context
	maxLineCoverage float64
	minLineCoverage float64
	moduleDir       string
	rootDir         string
	scriptsDir      string
	targetDir       string
	quality         *Quality
	caller          string
	scanChanged     bool
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

func NewProject(coverages ...float64) *Project {
	project := &Project{
		moduleDir:       moduleDir(),
		minLineCoverage: -1,
		maxLineCoverage: -1,
		quality: &Quality{
			LinterIssues: &LinterIssue{
				Detail: map[string]int{},
			},
		},
	}
	if len(coverages) == 1 {
		project.minLineCoverage = coverages[0]
		project.maxLineCoverage = 100
	} else if len(coverages) >= 2 {
		project.minLineCoverage = coverages[0]
		project.maxLineCoverage = coverages[1]
	}
	if project.minLineCoverage > 0 && project.minLineCoverage >= project.maxLineCoverage {
		log.Fatalf(color.RedString("invalid coverage range %f ~ %f", project.minLineCoverage, project.maxLineCoverage))
	}
	project.setup()
	return project
}

func (p *Project) ModuleDir() string {
	return p.moduleDir
}

func (p *Project) RootDir() string {
	return p.rootDir
}

func (p *Project) TargetDir() string {
	return p.targetDir
}

func (p *Project) Quality() *Quality {
	return p.quality
}

func (p *Project) WithCtx(ctx context.Context) *Project {
	p.ctx = ctx
	return p
}

func (p *Project) Ctx() context.Context {
	if p.ctx == nil {
		p.ctx = context.Background()
	}
	return p.ctx
}

func (p *Project) setup() {
	hookMap := map[string]string{
		"commit-msg": messageHook,
		"pre-push":   pushHook,
	}
	// setup folders
	p.targetDir = filepath.Join(p.moduleDir, buildTarget)
	p.scriptsDir = filepath.Join(p.moduleDir, scriptDir)
	os.MkdirAll(p.targetDir, os.ModePerm)
	os.MkdirAll(p.scriptsDir, os.ModePerm)
	p.rootDir = ProjectRoot(p.moduleDir)
	// setup caller
	pcs := make([]uintptr, 10)
	n := runtime.Callers(0, pcs)
	pcs = pcs[:n]
	frames := runtime.CallersFrames(pcs)
	for {
		frame, more := frames.Next()
		if !more {
			break
		}
		for _, v := range hookMap {
			if strings.HasSuffix(frame.File, v) {
				p.caller = v
				break
			}
		}
	}

	// setup hooks
	for h, c := range hookMap {
		h = filepath.Join(p.RootDir(), ".git", "hooks", h)
		c = filepath.Join(p.RootDir(), scriptDir, c)
		if _, err := os.Stat(c); err != nil {
			log.Fatalln(color.RedString("can not find %s, run command 'gbt githook' to initialize the hook", c))
		}
		command := fmt.Sprintf(scriptLine, c)
		if lines, err := os.ReadFile(h); err != nil || !strings.Contains(string(lines), command) {
			if f, err := os.Create(c); err == nil {
				f.WriteString("#!/bin/sh\n\n")
				f.WriteString(fmt.Sprintf("go run %s $1 $2\n", c))
				f.Close()
			} else {
				log.Fatalln(color.RedString("failed to generate %s", h))
			}
		}
	}
}

func (p *Project) processTestResult() {
	file, err := os.Open(filepath.Join(p.targetDir, rawTestReport))
	if err != nil {
		log.Fatalln(color.RedString("failed to open the file %v \n", filepath.Join(p.targetDir, rawTestReport)))
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
	p.quality.Tests = len(testSet)

	mc, err := os.Open(filepath.Join(p.targetDir, methodCoverageReport))
	if err != nil {
		log.Fatalln(color.RedString("failed to open the file %v", filepath.Join(p.targetDir, methodCoverageReport)))
	}
	defer mc.Close()
	testedMethod := 0
	scanner = bufio.NewScanner(mc)
	for scanner.Scan() {
		text := scanner.Text()
		items := strings.Fields(text)
		coverage, _ := strconv.ParseFloat(strings.TrimRight(items[2], "%"), 64)
		coverage /= 100
		if strings.EqualFold(items[0], "total:") {
			p.quality.Coverage.Line = coverage
		} else {
			p.quality.Methods++
			if coverage > 0 {
				testedMethod++
			}
		}
	}
	p.quality.Coverage.Method = math.Floor(float64(testedMethod)/float64(p.quality.Methods)*1000) / 1000
}

func (p *Project) Clean() *Project {
	log.Println("clean build directory ......")
	if err := os.RemoveAll(p.targetDir); err != nil {
		log.Fatalln(color.RedString("failed to delete %s\n", p.targetDir))
	}
	return p
}

// Test run the test with -race, -cover, -fuzz and -bench.
func (p *Project) Test(args ...string) *Project {
	log.Println("run unit test ......")
	os.Chdir(p.moduleDir)
	os.MkdirAll(p.targetDir, os.ModePerm)
	params := []string{"test", "-v", "-json", "-coverprofile", filepath.Join(p.targetDir, lineCoverageReport), "./..."}
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
	}
	os.WriteFile(filepath.Join(p.targetDir, rawTestReport), out, os.ModePerm)
	//  go tool cover -func ./targetDir/coverage.data
	params = []string{"tool", "cover", "-func", filepath.Join(p.targetDir, lineCoverageReport)}
	out, _ = exec.Command("go", params...).CombinedOutput()
	os.WriteFile(filepath.Join(p.targetDir, methodCoverageReport), out, os.ModePerm)
	p.processTestResult()
	log.Println(color.CyanString("total tests :%d, line coverage: %f, method coverage %f", p.quality.Tests, p.quality.Coverage.Line, p.quality.Coverage.Method))
	return p
}

// Build walk from module directory and run build command for each executable
// and place the executable at ${project_root}/bin; in case there are more than one executable.
func (p *Project) Build(files ...string) *Project {
	targetFiles := files
	if len(targetFiles) == 0 {
		targetFiles = append(targetFiles, "main.go")
	}
	log.Println("build project ......")
	os.MkdirAll(p.targetDir, os.ModePerm)
	filepath.Walk(p.moduleDir, func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		for _, t := range targetFiles {
			if strings.EqualFold(info.Name(), t) {
				if output, err := exec.Command("go", "build", "-o", p.targetDir, path).CombinedOutput(); err != nil {
					fmt.Println(string(output))
				}
			}
		}
		return nil
	})
	return p
}

func (p *Project) Scan(args ...string) *Project {
	log.Println("scan source code ......")
	if strings.EqualFold(p.caller, messageHook) {
		p.scanChanged = true
	}
	linter.Scan(p)
	data, _ := json.Marshal(p.quality)
	var prettyJSON bytes.Buffer
	json.Indent(&prettyJSON, data, "", "\t")
	os.WriteFile(filepath.Join(p.scriptsDir, quality), prettyJSON.Bytes(), os.ModePerm)
	if strings.EqualFold(p.caller, pushHook) && len(args) > 1 {
		p.pushGard(args...)
	}

	return p
}

func (p *Project) pushGard(revs ...string) {
}

func ProjectRoot(dir string) string {
	for dir != string(os.PathSeparator) {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir
		} else {
			dir = filepath.Dir(dir)
		}
	}
	if dir == string(os.PathSeparator) {
		fmt.Printf("project %s is not in the git repository\n", dir)
		dir = ""
	}
	return dir
}
