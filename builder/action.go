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
)

type Action interface {
	Do(ctx context.Context, cmd *Command) error
}

type actionFunc func(ctx context.Context, cmd *Command) error

func (action actionFunc) Do(ctx context.Context, cmd *Command) error {
	cmd.stack = append(cmd.stack, fmt.Sprintf("%T", action))
	ctxFlags, _ := ctx.Value(CtxKeyRunFlags).(map[string]bool)
	var flags []string
	for _, f := range cmd.Flags {
		if ctxFlags[f] {
			flags = append(flags, f)
		}
	}
	cmd.Flags = flags
	if len(flags) > 0 {
		log.Printf("Run %s with flags: %s\n", cmd.Name, strings.Join(flags, ","))
	}

	return action(ctx, cmd)
}

var _ Action = (*actionFunc)(nil)

var builderFunc actionFunc = func(ctx context.Context, cmd *Command) error {
	log.Println("Creating project build file")
	return infra.SetupBuilder(GetBuilder(ctx).ScriptDir()) //nolint:wrapcheck
}

var gitHookFunc actionFunc = func(ctx context.Context, cmd *Command) error {
	// value, ok := ctx.Value(GenHook).(bool)
	showMsg := cmd.Name == "gitHook"
	builder := GetBuilder(ctx)
	err := infra.GenGitHooks(builder.GitHome(), builder.ScriptDir())
	if err != nil {
		err = fmt.Errorf("failed to setup hook:%w", err)
	} else if showMsg {
		log.Println("git hooks are setup successfully")
	}
	return err
}

var createDirFunc actionFunc = func(ctx context.Context, cmd *Command) error {
	var dir string
	builder := GetBuilder(ctx)
	switch cmd.Name {
	case "gitHook", "builder":
		dir = builder.ScriptDir()
	case "lint", "test", "build":
		dir = builder.TargetDir()
	}
	if len(dir) < 1 {
		return nil
	}
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		err = fmt.Errorf("failed to create dir:%w", err)
	}
	return err
}

var cleanFunc actionFunc = func(ctx context.Context, cmd *Command) error {
	builder := GetBuilder(ctx)
	flags := append([]string{"clean"}, cmd.Flags...)
	output, err := exec.Command("go", flags...).CombinedOutput()
	msg := string(output)
	if err != nil {
		msg = color.RedString(string(output))
		log.Println(msg)
		return err //nolint:wrapcheck
	}
	log.Println(msg)
	log.Printf("Clean directory %s \n", GetBuilder(ctx).TargetDir())
	err = filepath.WalkDir(builder.TargetDir(), func(path string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() {
			err = os.Remove(path)
			if err != nil {
				log.Println(color.YellowString("failed to delete %s:%s", path, err.Error()))
			}
		} else if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err //nolint:wrapcheck
	})
	if err != nil {
		return fmt.Errorf("failed to delete %s :%w", builder.TargetDir(), err)
	}
	return err //nolint:wrapcheck

}

var commitMsgFunc actionFunc = func(ctx context.Context, cmd *Command) error {
	builder := GetBuilder(ctx)
	return infra.CommitMsg(string(builder.buildOption.MsgPattern)) //nolint:wrapcheck
}

var lintFunc actionFunc = func(ctx context.Context, cmd *Command) error {
	builder := GetBuilder(ctx)
	if builder.IsCommitHook() {
		cmd.Flags = append(cmd.Flags, "--fix", "-n")
	}
	return infra.LintScan(builder.RootDir(), builder.TargetDir(), cmd.Flags, builder.IsCommitHook()) //nolint:wrapcheck
}

var testFunc actionFunc = func(ctx context.Context, cmd *Command) error {
	builder := GetBuilder(ctx)
	err := os.Chdir(builder.RootDir())
	if err != nil {
		return fmt.Errorf("failed to change directory:%w", err)
	}
	params := []string{"test", "-v", "-coverprofile", filepath.Join(builder.TargetDir(), testCoverOut), "./..."}
	// @todo for test parameter
	// if len(args) > 0 {
	//	params = append(params, args...)
	//}

	testCmd := exec.Command("go", params...)
	stdout, err := testCmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("test failed:%w", err)

	}
	err = testCmd.Start()
	if err != nil {
		return fmt.Errorf("test failed:%w", err)
	}
	scanner := bufio.NewScanner(stdout)
	// ok  	github.com/kcmvp/gob/builder	0.155s	coverage: 16.9% of statements
	pkr := regexp.MustCompile(`\sok\s+\S+\s+\S+s\s+coverage:\s+\S+% of statements`)
	// ?   	github.com/kcmvp/gob/infra	[no test files]
	ntr := regexp.MustCompile(`\s+\S+\s+\[no test files\]`)
	pkgCoverage := map[string]string{}
	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case strings.Contains(line, "[build failed]"):
			line = color.RedString(line)
			err = errors.New("build failed") //nolint
		case strings.Contains(line, "--- FAIL:"), strings.Contains(line, "FAIL"):
			line = color.RedString(line)
			err = errors.New("test failure") //nolint
		case pkr.MatchString(line):
			line = color.GreenString(line)
			find := strings.Fields(pkr.FindString(line))
			pkgCoverage[find[1]] = find[4]
		case ntr.MatchString(line):
			line = color.YellowString(line)
			find := strings.Fields(ntr.FindString(line))
			pkgCoverage[find[0]] = "-"
		}
		log.Println(line)
	}
	if err != nil {
		return fmt.Errorf("test failed:%w", err)
	}

	data, err := json.MarshalIndent(&pkgCoverage, "", " ")
	if err != nil {
		return fmt.Errorf("failed to marshal package coverage report:%w", err)
	}
	err = os.WriteFile(filepath.Join(builder.TargetDir(), testPackageCover), data, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to save package coverage report:%w", err)
	}
	//  go tool cover -func ./targetDir/coverage.data
	fileCover := filepath.Join(builder.TargetDir(), testCoverReport)
	params = []string{"tool", "cover", "-html", filepath.Join(builder.TargetDir(), testCoverOut), "-o", fileCover}
	out, err := exec.Command("go", params...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s:%w", string(out), err)
	}
	log.Printf("coverage report is generated at %s \n", fileCover)
	return err //nolint:wrapcheck
}

var buildFunc actionFunc = func(ctx context.Context, cmd *Command) error {
	// @todo multiple main.go
	builder := GetBuilder(ctx)
	var targetFiles []string
	if len(targetFiles) == 0 {
		targetFiles = append(targetFiles, "main.go")
	}
	log.Println("build project ......")
	err := filepath.WalkDir(builder.RootDir(), func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}
		for _, t := range targetFiles {
			if strings.EqualFold(d.Name(), t) {
				if output, err := exec.Command("go", "build", "-o", builder.TargetDir(), path).CombinedOutput(); err != nil { //nolint
					log.Println(string(output))
					return err //nolint
				}
			}
		}
		return nil
	})

	if err != nil {
		err = fmt.Errorf("failed to build proejct:%w", err)
	}
	return err
}
