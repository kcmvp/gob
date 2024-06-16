package command

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"github.com/kcmvp/gob/cmd/gbc/artifact" //nolint
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
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
		{A: "after_lint", B: nolintReport},
	}
}
func (a Action) String() string {
	return fmt.Sprintf("%s: %s", a.A, a.A)
}

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
	return artifact.PtyCmdOutput(testCmd, "start test", true, nil)
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

func nolintReport(_ *cobra.Command, _ ...string) error {
	_ = os.Chdir(artifact.CurProject().Root())
	reg := regexp.MustCompile(`//\s*nolint`)
	data, _ := exec.Command("go", "list", "-f", `{{.Dir}}:{{join .GoFiles " "}}`, `./...`).CombinedOutput()
	scanner := bufio.NewScanner(bytes.NewBuffer(data))
	var ignoreList []lo.Tuple3[string, int, int]
	var maxLength, fOverAll, lOverAll int
	for scanner.Scan() {
		line := scanner.Text()
		items := strings.Split(line, ":")
		for _, file := range strings.Split(items[1], " ") {
			var fIgnore, lIgnore int
			abs := filepath.Join(items[0], file)
			relative := strings.TrimPrefix(abs, artifact.CurProject().Root()+"/")
			maxLength = lo.If(maxLength > len(relative), maxLength).Else(len(relative))
			source, _ := os.ReadFile(abs)
			sourceScanner := bufio.NewScanner(bytes.NewBuffer(source))
			for sourceScanner.Scan() {
				line = sourceScanner.Text()
				if reg.MatchString(line) {
					if strings.HasPrefix(strings.TrimPrefix(line, " "), "//") {
						fIgnore++
						lIgnore++
					} else {
						lIgnore++
					}
				}
			}
			fOverAll += fIgnore
			lOverAll += lIgnore
			if fIgnore+lIgnore > 0 {
				ignoreList = append(ignoreList, lo.Tuple3[string, int, int]{A: relative, B: fIgnore, C: lIgnore})
			}
		}
	}
	maxLength += 5
	if fOverAll+lOverAll > 0 {
		color.Yellow("[Lint Report]:file level ignores: %d, line level ignores: %d", fOverAll, lOverAll)
		slices.SortFunc(ignoreList, func(a, b lo.Tuple3[string, int, int]) int {
			return b.C - a.C
		})
		// Open a file for writing
		format := fmt.Sprintf("%%-%ds L:%%-10d F:%%-5d\n", maxLength)
		file, _ := os.Create(filepath.Join(artifact.CurProject().Target(), "lint-ignore.log"))
		defer file.Close()
		for _, line := range ignoreList {
			_, _ = file.WriteString(fmt.Sprintf(format, line.A, line.C, line.B))
		}
	}

	return nil
}
