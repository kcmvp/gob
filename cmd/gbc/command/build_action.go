package command

import (
	"errors"
	"fmt"
	"github.com/kcmvp/gob/cmd/gbc/artifact" //nolint
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/samber/lo" //nolint
	"github.com/spf13/cobra"
)

type (
	Execution func(cmd *cobra.Command, args ...string) error
	Action    lo.Tuple3[string, Execution, string]
)

func buildActions() []Action {
	return []Action{
		{A: "build", B: buildAction, C: "build all main methods which in main package and name the binary as file name"},
		{A: "clean", B: cleanAction, C: "clean project target folder"},
		{A: "test", B: testAction, C: "test the project and generate coverage report in target folder"},
		{A: "after_test", B: coverReport},
	}
}
func (a Action) String() string {
	return fmt.Sprintf("%s: %s", a.A, a.A)
}

//func setupActions() []Action {
//	return []Action{
//		{A: "version", B: setupVersion},
//	}
//}

func beforeExecution(cmd *cobra.Command, arg string) error {
	if action, ok := lo.Find(buildActions(), func(action Action) bool {
		return action.A == fmt.Sprintf("before_%s", arg)
	}); ok {
		return action.B(cmd, arg)
	}
	return nil
}

func afterExecution(cmd *cobra.Command, arg string) {
	if action, ok := lo.Find(buildActions(), func(action Action) bool {
		return action.A == fmt.Sprintf("after_%s", arg)
	}); ok {
		action.B(cmd, arg) //nolint
	}
}

func execute(cmd *cobra.Command, arg string) error {
	beforeExecution(cmd, arg) //nolint
	var err error
	if plugin, ok := lo.Find(artifact.CurProject().Plugins(), func(plugin artifact.Plugin) bool {
		return plugin.Alias == arg
	}); ok {
		err = plugin.Execute()
	} else if action, ok := lo.Find(buildActions(), func(action Action) bool {
		return action.A == arg
	}); ok {
		err = action.B(cmd, arg)
	} else {
		return fmt.Errorf(color.RedString("can not find command %s", arg))
	}
	if err != nil {
		return err
	}
	afterExecution(cmd, arg)
	return nil
}

func validBuilderArgs() []string {
	builtIn := lo.Map(lo.Filter(buildActions(), func(item Action, _ int) bool {
		return !strings.Contains(item.A, "_")
	}), func(action Action, _ int) string {
		return action.A
	})
	lo.ForEach(artifact.CurProject().Plugins(), func(item artifact.Plugin, _ int) {
		if !lo.Contains(builtIn, item.Alias) {
			builtIn = append(builtIn, item.Alias)
		}
	})
	return builtIn
}

func buildAction(_ *cobra.Command, _ ...string) error {
	bm := map[string]string{}
	for _, mainFile := range artifact.CurProject().MainFiles() {
		binary := strings.TrimSuffix(filepath.Base(mainFile), ".go")
		if f, exists := bm[binary]; exists {
			return fmt.Errorf("file %s has already built as %s, please rename %s", f, binary, mainFile)
		}
		output := filepath.Join(artifact.CurProject().Target(), binary)
		versionFlag := fmt.Sprintf("-X 'main.buildVersion=%s'", artifact.Version())
		if msg, err := exec.Command("go", "build", "-o", output, "-ldflags", versionFlag, mainFile).CombinedOutput(); err != nil { //nolint
			return errors.New(color.RedString(string(msg)))
		}
		fmt.Printf("Build project successfully %s\n", output)
		bm[binary] = output
	}
	if len(bm) == 0 {
		color.Yellow("Can't find main methods")
	}
	return nil
}

func cleanAction(_ *cobra.Command, _ ...string) error {
	// clean target folder
	os.RemoveAll(artifact.CurProject().Target())
	os.Mkdir(artifact.CurProject().Target(), os.ModePerm) //nolint errcheck
	fmt.Println("Clean target folder successfully !")
	// clean cache
	args := []string{"clean"}
	_, err := exec.Command("go", args...).CombinedOutput()
	if err != nil {
		color.Red("failed to clean the project cache %s", err.Error())
	}
	return err
}

func testAction(_ *cobra.Command, _ ...string) error {
	coverProfile := fmt.Sprintf("-coverprofile=%s/cover.out", artifact.CurProject().Target())
	testCmd := exec.Command("go", []string{"test", "-v", coverProfile, "./..."}...) //nolint
	return artifact.StreamCmdOutput(testCmd, "test")
}

func coverReport(_ *cobra.Command, _ ...string) error {
	target := artifact.CurProject().Target()
	_, err := os.Stat(filepath.Join(target, "cover.out"))
	if err == nil {
		if _, err = exec.Command("go", []string{"tool", "cover", fmt.Sprintf("-html=%s/cover.out", target), fmt.Sprintf("-o=%s/cover.html", target)}...).CombinedOutput(); err == nil { //nolint
			fmt.Printf("Coverage report is generated at %s/cover.html \n", target)
			return nil
		}
		return fmt.Errorf(color.RedString("Failed to generate coverage report %s", err.Error()))
	}
	return fmt.Errorf(color.RedString("Failed to generate coverage report %s", err.Error()))
}

//func setupVersion(_ *cobra.Command, _ ...string) error {
//	infra := filepath.Join(artifact.CurProject().Root(), "infra")
//	if _, err := os.Stat(infra); err != nil {
//		os.Mkdir(infra, 0700) // nolint
//	}
//	ver := filepath.Join(infra, "version.go")
//	if _, err := os.Stat(ver); err != nil {
//		data, _ := resources.ReadFile(filepath.Join(resourceDir, "version.tmpl"))
//		os.WriteFile(ver, data, 0666) //nolint
//	}
//	color.GreenString("version file is generated at infra/version.go")
//	return nil
//}
