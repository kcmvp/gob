package builder

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/fatih/color"
	"github.com/kcmvp/gob/infra"
	"github.com/looplab/fsm"
)

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
	infra.CheckError(err, "failed to change directory")
	CleanResult(Test)
	params := []string{"test", "-v", "-coverprofile", filepath.Join(instance.TargetDir(), testCoverOut), "./..."}
	// @todo for test parameter
	// if len(args) > 0 {
	//	params = append(params, args...)
	//}

	testCmd := exec.Command("go", params...)
	stdout, err := testCmd.StdoutPipe()
	infra.CheckError(err, "runs into error")
	err = testCmd.Start()
	infra.CheckError(err, "Failed to start test")
	scanner := bufio.NewScanner(stdout)
	// ok  	github.com/kcmvp/gob/builder	0.155s	coverage: 16.9% of statements
	pkr := regexp.MustCompile(`\sok\s+\S+\s+\S+s\s+coverage:\s+\S+% of statements`)
	// ?   	github.com/kcmvp/gob/infra	[no test files]
	ntr := regexp.MustCompile(`\s+\S+\s+\[no test files\]`)
	pkgCoverage := map[string]string{}
	for scanner.Scan() {
		m := scanner.Text()
		switch {
		case strings.Contains(m, "[build failed]"):
			m = color.RedString(m)
			err = errors.New("build failed") //nolint
		case strings.Contains(m, "--- FAIL:"):
			m = color.RedString(m)
			err = errors.New("test failure") //nolint
		case pkr.MatchString(m):
			m = color.GreenString(m)
			find := strings.Fields(pkr.FindString(m))
			pkgCoverage[find[1]] = find[4]
		case ntr.MatchString(m):
			m = color.YellowString(m)
			find := strings.Fields(ntr.FindString(m))
			pkgCoverage[find[0]] = "-"
		}
		log.Println(m)
	}

	if err != nil {
		event.Cancel(err)
	} else {
		data, err := json.MarshalIndent(&pkgCoverage, "", " ")
		infra.CheckError(err, "Failed to marshal package coverage report")
		err = os.WriteFile(filepath.Join(instance.targetDir, testPackageCover), data, os.ModePerm)
		infra.CheckError(err, "Failed to save package coverage report")
	}
	//  go tool cover -func ./targetDir/coverage.data
	fileCover := filepath.Join(instance.TargetDir(), testCoverReport)
	params = []string{"tool", "cover", "-html", filepath.Join(instance.TargetDir(), testCoverOut), "-o", fileCover}
	out, err := exec.Command("go", params...).CombinedOutput()
	infra.CheckError(err, string(out))
	log.Printf("coverage report is generated at %s \n", fileCover)
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
	infra.CheckError(err, "Failed to build project")
}
