package builder

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

var createDirCallback fsm.Callback = func(ctx context.Context, event *fsm.Event) {
	var dir string
	builder := GetBuilder(ctx)
	switch event.Dst {
	case string(GenGitHook), string(GenBuilder):
		dir = builder.ScriptDir()
	case string(Lint), string(Test), string(Build):
		dir = builder.TargetDir()
	}
	if len(dir) < 1 {
		return
	}
	err := os.MkdirAll(dir, os.ModePerm)
	cancelEvent(event, err, "failed to create dir")
}

var cleanAllCallback fsm.Callback = func(ctx context.Context, event *fsm.Event) {
	builder := GetBuilder(ctx)
	log.Printf("clean directory %s \n", GetBuilder(ctx).TargetDir())
	err := os.RemoveAll(GetBuilder(ctx).TargetDir())
	cancelEvent(event, err, fmt.Sprintf("failed to delete %s\n", builder.TargetDir()))
}

var afterCleanAllCallback fsm.Callback = func(ctx context.Context, event *fsm.Event) {
	log.Println("clean success")
}

var preCommitCallback fsm.Callback = func(ctx context.Context, event *fsm.Event) {
}

var commitMsgCallback fsm.Callback = func(ctx context.Context, event *fsm.Event) {
	builder := GetBuilder(ctx)
	infra.CommitMsg(string(builder.buildOption.MsgPattern))
}

var prePushHookCallback fsm.Callback = func(ctx context.Context, event *fsm.Event) {
	log.Println("validate code quality for push")
}

var lintCallback fsm.Callback = func(ctx context.Context, event *fsm.Event) {
	builder := GetBuilder(ctx)
	isPreCommitHooK := preCommitHook == builder.hook
	v, _ := ctx.Value(ScanAll).(bool)
	v = v && !isPreCommitHooK
	builder.cleanOutput(Lint)
	infra.LintScan(builder.RootDir(), builder.TargetDir(), v, isPreCommitHooK)
}

var testCallback fsm.Callback = func(ctx context.Context, event *fsm.Event) {
	builder := GetBuilder(ctx)
	err := os.Chdir(builder.RootDir())
	cancelEvent(event, err, "failed to change directory")
	builder.cleanOutput(Test)
	params := []string{"test", "-v", "-coverprofile", filepath.Join(builder.TargetDir(), testCoverOut), "./..."}
	// @todo for test parameter
	// if len(args) > 0 {
	//	params = append(params, args...)
	//}

	testCmd := exec.Command("go", params...)
	stdout, err := testCmd.StdoutPipe()
	cancelEvent(event, err, "runs into error")
	err = testCmd.Start()
	cancelEvent(event, err, "runs into error")
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
		cancelEvent(event, err, "failed to marshal package coverage report")
		err = os.WriteFile(filepath.Join(builder.TargetDir(), testPackageCover), data, os.ModePerm)
		cancelEvent(event, err, "failed to save package coverage report")
	}
	//  go tool cover -func ./targetDir/coverage.data
	fileCover := filepath.Join(builder.TargetDir(), testCoverReport)
	params = []string{"tool", "cover", "-html", filepath.Join(builder.TargetDir(), testCoverOut), "-o", fileCover}
	out, err := exec.Command("go", params...).CombinedOutput()
	cancelEvent(event, err, string(out))
	log.Printf("coverage report is generated at %s \n", fileCover)
}

var buildCallback fsm.Callback = func(ctx context.Context, event *fsm.Event) {
	// @todo multiple main.go
	builder := GetBuilder(ctx)
	var targetFiles []string
	if len(targetFiles) == 0 {
		targetFiles = append(targetFiles, "main.go")
	}
	log.Println("build project ......")
	err := filepath.Walk(builder.TargetDir(), func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		for _, t := range targetFiles {
			if strings.EqualFold(info.Name(), t) {
				if output, err := exec.Command("go", "build", "-o", builder.TargetDir(), path).CombinedOutput(); err != nil { //nolint
					log.Println(string(output))
					return err //nolint
				}
			}
		}
		return nil
	})
	cancelEvent(event, err, "failed to build project")
}
