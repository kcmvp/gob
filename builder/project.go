package builder

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/fatih/color"
	"github.com/kcmvp/gob/infra"
)

const (
	lineCoverageReport   = "line.data"
	methodCoverageReport = "method.data"
	rawTestReport        = "test.data"
	scriptDir            = "scripts"
	targetDir            = "target"
	coverage             = "coverage.json"
)

type project struct {
	moduleDir string
	scriptDir string
	targetDir string
}

type TestCase struct {
	Package string
	Test    string
	Action  string
	Output  string
	Elapsed float64
}

func newProject(root string) *project {
	p := &project{
		moduleDir: root,
		targetDir: filepath.Join(root, targetDir),
		scriptDir: filepath.Join(root, scriptDir),
	}
	err := os.MkdirAll(p.targetDir, os.ModePerm)
	FatalIfError(err)
	err = os.MkdirAll(p.scriptDir, os.ModePerm)
	FatalIfError(err)
	return p
}

func (project *project) ModuleDir() string {
	return project.moduleDir
}

func (project *project) TargetDir() string {
	return project.targetDir
}

func (project *project) GirDir() string {
	return filepath.Join(project.ModuleDir(), ".git")
}

func (project *project) ScriptDir() string {
	return project.scriptDir
}

func (project *project) clean() {
	log.Printf("clean directory %s \n", project.targetDir)
	if err := os.RemoveAll(project.targetDir); err != nil {
		log.Fatalln(color.RedString("failed to delete %s\n", project.targetDir))
	} else {
		if err := os.MkdirAll(project.targetDir, os.ModePerm); err != nil {
			log.Fatalln(color.RedString("failed to create directory %s, err : %s", project.targetDir, err.Error()))
		}
	}
}

// Test run the test with -race, -cover, -fuzz and -bench.
func (project *project) test(args ...string) {
	os.Chdir(project.moduleDir)
	os.MkdirAll(project.targetDir, os.ModePerm)
	params := []string{"test", "-v", "-json", "-coverprofile", filepath.Join(project.targetDir, lineCoverageReport), "./..."}
	if len(args) > 0 {
		params = append(params, args...)
	}
	out, _ := exec.Command("go", params...).CombinedOutput()

	if err := os.WriteFile(filepath.Join(project.targetDir, rawTestReport), out, os.ModePerm); err != nil {
		log.Fatalln(color.RedString("failed to generate coverage report:%s", err.Error()))
	}
	//  go tool cover -func ./targetDir/coverage.data
	fileCover := filepath.Join(project.targetDir, "cover_file.html")
	params = []string{"tool", "cover", "-html", filepath.Join(project.targetDir, lineCoverageReport), "-o", fileCover}
	_, err := exec.Command("go", params...).CombinedOutput()
	checkError(err)
	log.Printf("coverage report is generated at %s \n", fileCover)
}

// Build walk from module directory and run build command for each executable
// and place the executable at ${project_root}/bin; in case there are more than one executable.
func (project *project) build(files ...string) *project {
	targetFiles := files
	if len(targetFiles) == 0 {
		targetFiles = append(targetFiles, "main.go")
	}
	log.Println("build project ......")
	os.MkdirAll(project.targetDir, os.ModePerm) //nolint:errcheck
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

func (project *project) coverage(keepInGit bool) {
	cover := filepath.Join(project.targetDir, coverage)
	// if keepInGit {
	//	cover = filepath.Join(project.moduleDir, coverage)
	//}
	file, err := os.Open(filepath.Join(project.targetDir, rawTestReport))
	if err != nil {
		log.Fatalln(color.RedString("failed to open the file %v \n", filepath.Join(project.targetDir, rawTestReport)))
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)

	report := infra.Report{
		Packages: map[string]string{},
	}
	testSet := map[string]bool{}
	for scanner.Scan() {
		text := scanner.Text()
		c := TestCase{}
		json.Unmarshal([]byte(text), &c) //nolint:errcheck
		if len(c.Test) > 0 {
			testSet[c.Test] = true
		} else if len(c.Output) > 0 {
			if strings.Contains(c.Output, "no test files") {
				report.Packages[c.Package] = "-"
			} else if strings.HasPrefix(c.Output, "coverage:") {
				report.Packages[c.Package] = strings.Fields(c.Output)[1]
			}
		}
	}
	report.Tests = len(testSet)
	data, _ := json.MarshalIndent(report, "", " ")
	if os.WriteFile(cover, data, os.ModePerm) == nil {
		log.Printf("coverage report is generated at %s", cover)
	}
}

func checkError(err error) {
	if err != nil {
		log.Fatalln(color.RedString("runs into error: %s", err.Error()))
	}
}
