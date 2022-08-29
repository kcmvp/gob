package builder

import (
	"bufio"
	"context"
	"encoding/json"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/kcmvp/gob/infra"
	"github.com/looplab/fsm"
)

type TestCase struct {
	Package string
	Test    string
	Action  string
	Output  string
	Elapsed float64
}

var setupBuilderCallback fsm.Callback = func(ctx context.Context, event *fsm.Event) {
	log.Println("Creating project build file")
	infra.SetupBuilder(instance.scriptDir)
}

var createDirCallback fsm.Callback = func(ctx context.Context, event *fsm.Event) {
	var dir string
	switch event.Dst {
	case string(SetupGitHook), string(SetupBuilder):
		dir = instance.scriptDir
	case string(Lint), string(Test), string(Build):
		dir = instance.targetDir
	}
	if len(dir) < 1 {
		return
	}
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		log.Fatalln(color.RedString("Failed to create dir %s: %s", dir, err.Error()))
	}
}

var gitHookCallback fsm.Callback = func(ctx context.Context, event *fsm.Event) {
	v, ok := ctx.Value(GenHook).(bool)
	if err := infra.SetupHook(projectScriptDir, ok && v); err != nil {
		log.Fatalln(color.RedString("failed to setup hook: %s", err.Error()))
	}
}

var cleanCallback fsm.Callback = func(ctx context.Context, event *fsm.Event) {
	log.Printf("clean directory %s \n", instance.TargetDir())
	if err := os.RemoveAll(instance.TargetDir()); err != nil {
		log.Fatalln(color.RedString("failed to delete %s\n", instance.TargetDir()))
	}
}

var afterCleanCallback fsm.Callback = func(ctx context.Context, event *fsm.Event) {
	log.Println("clean success")
}

var preCommitCallback fsm.Callback = func(ctx context.Context, event *fsm.Event) {
}

var commitMsgCallback fsm.Callback = func(ctx context.Context, event *fsm.Event) {
	infra.CommitMsg(string(instance.buildOption.MsgPattern))
}

var prePushHookCallback fsm.Callback = func(ctx context.Context, event *fsm.Event) {
	log.Println("validate code quality for push")
}

var lintCallback fsm.Callback = func(ctx context.Context, event *fsm.Event) {
	isPreCommitHooK := preCommitHook == instance.hook
	v, _ := ctx.Value(ScanAll).(bool)
	v = v && !isPreCommitHooK
	infra.LintScan(instance.TargetDir(), v, isPreCommitHooK)
}

var testCallback fsm.Callback = func(ctx context.Context, event *fsm.Event) {
	err := os.Chdir(instance.root)
	checkError(err, "failed to change directory")
	CleanResult(Test)
	params := []string{"test", "-v", "-json", "-coverprofile", filepath.Join(instance.TargetDir(), testCoverOut), "./..."}
	// @todo for test parameter
	// if len(args) > 0 {
	//	params = append(params, args...)
	//}
	out, err := exec.Command("go", params...).CombinedOutput()
	checkError(err, string(out))

	if err := os.WriteFile(filepath.Join(instance.TargetDir(), testPackageOut), out, os.ModePerm); err != nil {
		log.Fatalln(color.RedString("failed to generate coverage report:%s", err.Error()))
	} else {
		log.Printf("test output is generated at %s", filepath.Join(instance.TargetDir(), testPackageOut))
	}
	//  go tool cover -func ./targetDir/coverage.data
	fileCover := filepath.Join(instance.TargetDir(), "cover_file.html")
	params = []string{"tool", "cover", "-html", filepath.Join(instance.TargetDir(), testCoverOut), "-o", fileCover}
	out, err = exec.Command("go", params...).CombinedOutput()
	checkError(err, string(out))
	log.Printf("coverage report is generated at %s \n", fileCover)
}

var afterTestCallback fsm.Callback = func(ctx context.Context, event *fsm.Event) {
	cover := filepath.Join(instance.TargetDir(), testCoverageJSON)
	file, err := os.Open(filepath.Join(instance.TargetDir(), testPackageOut))
	if err != nil {
		log.Fatalln(color.RedString("failed to open the file %v \n", filepath.Join(instance.TargetDir(), testPackageOut)))
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
	data, err := json.MarshalIndent(report, "", " ")
	checkError(err, "Failed to marshall json")
	if os.WriteFile(cover, data, os.ModePerm) == nil {
		log.Printf("coverage report is generated at %s", cover)
	}
}

var buildCallback fsm.Callback = func(ctx context.Context, event *fsm.Event) {
	// @todo multiple main.go
	var targetFiles []string
	if len(targetFiles) == 0 {
		targetFiles = append(targetFiles, "main.go")
	}
	log.Println("build project ......")
	err := filepath.Walk(instance.TargetDir(), func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		for _, t := range targetFiles {
			if strings.EqualFold(info.Name(), t) {
				if output, err := exec.Command("go", "build", "-o", instance.TargetDir(), path).CombinedOutput(); err != nil { //nolint
					log.Println(string(output))
					return err //nolint
				}
			}
		}
		return nil
	})
	checkError(err, "Failed to build project")
}

func checkError(err error, msg string) {
	if err != nil {
		log.Fatalln(color.RedString("%s: %s", msg, err.Error()))
	}
}
