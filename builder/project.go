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
)

const (
	lineCoverageReport   = "line.data"
	methodCoverageReport = "method.data"
	rawTestReport        = "test.data"
	quality              = "quality.json"
	GitHook              = "git"
	scriptDir            = "scripts"
	scriptLine           = "go run %s $1 $2 -%s \n"
	buildTarget          = "target"
)

type Quality struct {
	Methods  int
	Tests    int
	Coverage Coverage
	Issues   Issue
}

type Coverage struct {
	Method float64
	Line   float64
}

type Issue struct {
	Files   int
	Issues  int
	Linters map[string]int
}

type Project struct {
	ctx             context.Context
	maxLineCoverage float64
	minLineCoverage float64
	moduleDir       string
	rootDir         string
	scriptsDir      string
	targetDir       string
	quality         Quality
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
		quality: Quality{
			Issues: Issue{
				Linters: map[string]int{},
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
		log.Fatalf("invalid coverage range %f ~ %f", project.minLineCoverage, project.maxLineCoverage)
	}
	project.targetDir = filepath.Join(project.moduleDir, buildTarget)
	project.scriptsDir = filepath.Join(project.moduleDir, scriptDir)
	os.MkdirAll(project.targetDir, os.ModePerm)
	os.MkdirAll(project.scriptsDir, os.ModePerm)
	project.rootDir = ProjectRoot(project.moduleDir)
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

func (p *Project) Quality() Quality {
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

func (p *Project) Setup() *Project {
	if len(p.rootDir) > 0 {
		hookMap := map[string]string{
			"commit-msg": "message_hook.go",
			"pre-push":   "push_hook.go",
		}
		for h, c := range hookMap {
			h = filepath.Join(p.rootDir, ".git", h)
			c = filepath.Join(p.rootDir, scriptDir, c)
			if lines, err := os.ReadFile(h); err == nil {
				command := fmt.Sprintf(scriptLine, c, GitHook)
				if !strings.Contains(string(lines), command) {
					fmt.Printf("please delete %s and run command 'gbtc githook' to setup hook", h)
					os.Exit(1)
				}
			} else {
				fmt.Println("please run command 'gbtc githook' to setup project")
				os.Exit(1)
			}
		}
	}
	return p
}

func (p *Project) Clean() *Project {
	fmt.Println("clean build directory ......")
	if err := os.RemoveAll(p.targetDir); err != nil {
		fmt.Printf("failed to delete %s\n", p.targetDir)
		os.Exit(1)
	}
	return p
}

func process(p *Project) {
	file, err := os.Open(filepath.Join(p.targetDir, rawTestReport))
	if err != nil {
		fmt.Printf("failed to open the file %v \n", filepath.Join(p.targetDir, rawTestReport))
		os.Exit(1)
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
		fmt.Printf(fmt.Sprintf("failed to open the file %v", filepath.Join(p.targetDir, methodCoverageReport)))
		os.Exit(1)
	}
	defer mc.Close()
	testedMethod := 0
	scanner = bufio.NewScanner(mc)
	for scanner.Scan() {
		text := scanner.Text()
		items := strings.Fields(text)
		coverage, _ := strconv.ParseFloat(strings.TrimRight(items[2], "%"), 64)
		coverage = coverage / 100
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

// Test run the test with -race, -cover, -fuzz and -bench.
func (p *Project) Test(args ...string) *Project {
	if v, ok := p.Ctx().Value("event").(string); ok {
		if strings.HasPrefix(v, "message") {
			return p
		}
	}
	fmt.Println("run unit test ......")
	os.Chdir(p.moduleDir)
	os.MkdirAll(p.targetDir, os.ModePerm)
	params := []string{"test", "-v", "-json", "-coverprofile", filepath.Join(p.targetDir, lineCoverageReport), "./..."}
	if len(args) > 0 {
		params = append(params, args...)
	}
	out, err := exec.Command("go", params...).CombinedOutput()
	fmt.Println(string(out))
	if err != nil {
		os.Exit(1)
	}
	os.WriteFile(filepath.Join(p.targetDir, rawTestReport), out, os.ModePerm)
	//  go tool cover -func ./targetDir/coverage.data
	params = []string{"tool", "cover", "-func", filepath.Join(p.targetDir, lineCoverageReport)}
	out, _ = exec.Command("go", params...).CombinedOutput()
	os.WriteFile(filepath.Join(p.targetDir, methodCoverageReport), out, os.ModePerm)
	process(p)
	return p
}

// Build walk from module directory and run build command for each executable
// and place the executable at ${project_root}/bin; in case there are more than one executable.
func (p *Project) Build(files ...string) *Project {
	if v, _ := p.Ctx().Value("event").(string); len(v) > 0 {
		return p
	}
	targetFiles := files
	if len(targetFiles) == 0 {
		targetFiles = append(targetFiles, "main.go")
	}
	fmt.Println("build project ......")
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
	fmt.Println("scan source code ......")
	GolangCiLinter.Exec(p)
	return p
}

func (p *Project) Report() *Project {
	data, _ := json.Marshal(p.quality)
	var prettyJSON bytes.Buffer
	json.Indent(&prettyJSON, data, "", "\t")
	os.WriteFile(filepath.Join(p.scriptsDir, quality), prettyJSON.Bytes(), os.ModePerm)
	return p
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
		fmt.Printf("project %s is not in the git repository", dir)
		dir = ""
	}
	return dir
}
