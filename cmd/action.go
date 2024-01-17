package cmd

import (
	"errors"
	"fmt"
	"github.com/kcmvp/gob/shared" //nolint
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/kcmvp/gob/internal"
	"github.com/samber/lo" //nolint
	"github.com/spf13/cobra"
)

var (
	CleanCache     bool
	CleanTestCache bool
	CleanModCache  bool
)

const (
	cleanCacheFlag     = "cache"
	cleanTestCacheFlag = "testcache"
	cleanModCacheFlag  = "modcache"
)

type (
	Execution func(cmd *cobra.Command, args ...string) error
	Action    lo.Tuple2[string, Execution]
)

func builtinActions() []Action {
	return []Action{
		{A: "build", B: buildAction},
		{A: "clean", B: cleanAction},
		{A: "test", B: testAction},
		{A: "after_test", B: coverReport},
	}
}

func beforeExecution(cmd *cobra.Command, arg string) error {
	if action, ok := lo.Find(builtinActions(), func(action Action) bool {
		return action.A == fmt.Sprintf("before_%s", arg)
	}); ok {
		return action.B(cmd, arg)
	}
	return nil
}

func afterExecution(cmd *cobra.Command, arg string) error {
	if action, ok := lo.Find(builtinActions(), func(action Action) bool {
		return action.A == fmt.Sprintf("after_%s", arg)
	}); ok {
		return action.B(cmd, arg)
	}
	return nil
}

func execute(cmd *cobra.Command, arg string) error {
	err := beforeExecution(cmd, arg)
	if err != nil {
		return err
	}
	msg := fmt.Sprintf("Start %s project", arg)
	color.Cyan("%-20s ...... \n", msg)
	if plugin, ok := lo.Find(internal.CurProject().Plugins(), func(plugin internal.Plugin) bool {
		return plugin.Alias == arg
	}); ok {
		err = plugin.Execute()
	} else if action, ok := lo.Find(builtinActions(), func(action Action) bool {
		return action.A == arg
	}); ok {
		err = action.B(cmd, arg)
	} else {
		return fmt.Errorf("can not find command %s", arg)
	}
	if err == nil {
		return afterExecution(cmd, arg)
	}
	return err
}

func validBuilderArgs() []string {
	builtIn := lo.Map(lo.Filter(builtinActions(), func(item Action, _ int) bool {
		return !strings.Contains(item.A, "_")
	}), func(action Action, _ int) string {
		return action.A
	})
	lo.ForEach(internal.CurProject().Plugins(), func(item internal.Plugin, _ int) {
		if !lo.Contains(builtIn, item.Alias) {
			builtIn = append(builtIn, item.Alias)
		}
	})
	return builtIn
}

func buildAction(_ *cobra.Command, _ ...string) error {
	bm := map[string]string{}
	for _, mainFile := range internal.CurProject().MainFiles() {
		binary := strings.TrimSuffix(filepath.Base(mainFile), ".go")
		if f, exists := bm[binary]; exists {
			return fmt.Errorf("file %s has already built as %s, please rename %s", f, binary, mainFile)
		}
		output := filepath.Join(internal.CurProject().Target(), binary)
		versionFlag := fmt.Sprintf("-X '%s/infra.buildVersion=%s'", internal.CurProject().Module(), internal.Version())
		if _, err := exec.Command("go", "build", "-ldflags", versionFlag, "-o", output, mainFile).CombinedOutput(); err != nil { //nolint
			return errors.New(color.RedString("failed to build the project: %s", err.Error()))
		}
		fmt.Printf("Build %s to %s successfully\n", mainFile, output)
		bm[binary] = output
	}
	if len(bm) == 0 {
		color.Yellow("Can't find main methods")
	}
	return nil
}

func cleanAction(_ *cobra.Command, _ ...string) error {
	// clean target folder
	os.RemoveAll(internal.CurProject().Target())
	os.Mkdir(internal.CurProject().Target(), os.ModePerm) //nolint errcheck
	fmt.Println("Clean target folder successfully !")
	// clean cache
	args := []string{"clean"}
	if CleanCache {
		args = append(args, fmt.Sprintf("--%s", cleanCacheFlag))
	}
	if CleanTestCache {
		args = append(args, fmt.Sprintf("--%s", cleanTestCacheFlag))
	}
	if CleanModCache {
		args = append(args, fmt.Sprintf("--%s", cleanModCacheFlag))
	}
	_, err := exec.Command("go", args...).CombinedOutput()
	if len(args) > 1 && err == nil {
		fmt.Println("Clean cache successfully !")
	}
	return nil
}

func testAction(_ *cobra.Command, args ...string) error {
	coverProfile := fmt.Sprintf("-coverprofile=%s/cover.out", internal.CurProject().Target())
	testCmd := exec.Command("go", []string{"test", "-v", coverProfile, "./..."}...) //nolint
	return shared.StreamCmdOutput(testCmd, fmt.Sprintf("%s/test.log", internal.CurProject().Target()))
}

func coverReport(_ *cobra.Command, _ ...string) error {
	target := internal.CurProject().Target()
	_, err := os.Stat(filepath.Join(target, "cover.out"))
	if err == nil {
		if _, err = exec.Command("go", []string{"tool", "cover", fmt.Sprintf("-html=%s/cover.out", target), fmt.Sprintf("-o=%s/cover.html", target)}...).CombinedOutput(); err == nil { //nolint
			fmt.Printf("Coverage report is generated at %s/cover.html \n", target)
			return nil
		} else {
			return fmt.Errorf(color.RedString("Failed to generate coverage report %s", err.Error()))
		}
	}
	return fmt.Errorf(color.RedString("Failed to generate coverage report %s", err.Error()))
}
