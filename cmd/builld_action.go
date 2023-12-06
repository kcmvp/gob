package cmd

import (
	"bufio"
	"fmt"
	"github.com/fatih/color"
	"github.com/kcmvp/gob/internal"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	cleanAction  = "clean"
	testAction   = "test"
	lintAction   = "lint"
	buildAction  = "build"
	TargetFolder = "target"
)

var blueMsg = color.New(color.FgRed)
var targetFolder = fmt.Sprintf("%s/target", internal.CurProject().Root())

type buildCmdFunc func(cmd *cobra.Command) error

func findMain(dir string) (string, error) {
	var mf string
	re := regexp.MustCompile(`func\s+main\s*\(\s*\)`)
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && dir != path {
			return filepath.SkipDir
		}
		if d.IsDir() || !strings.HasSuffix(d.Name(), ".go") || strings.HasSuffix(d.Name(), "_test.go") {
			return nil
		}
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if re.MatchString(line) {
				mf = path
				return filepath.SkipDir
			}
		}
		return scanner.Err()
	})
	return mf, err
}

var cleanFunc buildCmdFunc = func(cmd *cobra.Command) error {
	// clean target folder
	target := filepath.Join(internal.CurProject().Root(), TargetFolder)
	filepath.WalkDir(target, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path != target {
			if err = os.RemoveAll(path); err != nil {
				return err
			}
		}
		return nil
	})
	fmt.Println("Clean target folder successfully !")
	// clean cache
	args := []string{"clean"}
	if ok, _ := cmd.Flags().GetBool(CleanCacheFlag); ok {
		args = append(args, fmt.Sprintf("--%s", CleanCacheFlag))
	}
	if ok, _ := cmd.Flags().GetBool(CleanTestCacheFlag); ok {
		args = append(args, fmt.Sprintf("--%s", CleanTestCacheFlag))
	}
	if ok, _ := cmd.Flags().GetBool(CleanModCacheFlag); ok {
		args = append(args, fmt.Sprintf("--%s", CleanModCacheFlag))
	}
	_, err := exec.Command("go", args...).CombinedOutput()
	if len(args) > 1 && err == nil {
		fmt.Println("Clean cache successfully !")
	}
	return nil
}

var lintFunc buildCmdFunc = func(cmd *cobra.Command) error {
	return nil
}

var testFunc buildCmdFunc = func(cmd *cobra.Command) error {
	coverprofile := fmt.Sprintf("-coverprofile=%s/cover.out", targetFolder)
	testCmd := exec.Command("go", []string{"test", "-v", coverprofile, "./..."}...)
	err := streamOutput(testCmd, fmt.Sprintf("%s/test.log", targetFolder), "FAIL:")
	exec.Command("go", []string{"tool", "cover", fmt.Sprintf("-html=%s/cover.out", targetFolder), fmt.Sprintf("-o=%s/cover.html", targetFolder)}...).CombinedOutput()
	cc := lo.IfF(err != nil, func() color.Attribute {
		return color.FgYellow
	}).Else(color.FgGreen)
	c := color.New(cc)
	c.Printf("Test report is generated at %s/test.log \n", targetFolder)
	c.Printf("Coverage report is generated at %s/cover.html \n", targetFolder)
	return err
}

var buildFunc buildCmdFunc = func(cmd *cobra.Command) error {
	dirs, err := internal.FindGoFilesByPkg("main")
	if err != nil {
		return err
	}
	bm := map[string]string{}
	for _, dir := range dirs {
		mf, err := findMain(dir)
		if err != nil {
			return err
		}
		if len(mf) > 0 {
			// action
			binary := strings.TrimSuffix(filepath.Base(mf), ".go")
			if f, exists := bm[binary]; exists {
				return fmt.Errorf("file %s has already built as %s, please rename %s", f, binary, mf)
			}
			output := filepath.Join(internal.CurProject().Root(), TargetFolder, binary)
			if _, err := exec.Command("go", "build", "-o", output, mf).CombinedOutput(); err != nil { //nolint
				return err
			} else {
				fmt.Printf("Build %s to %s successfully\n", mf, output)
				bm[binary] = output
			}
		} else {
			color.Yellow("Can not find main function in package %s", dir)
		}
	}
	return nil
}

var buildActions = []lo.Tuple2[string, buildCmdFunc]{
	lo.T2(cleanAction, cleanFunc),
	lo.T2(testAction, testFunc),
	lo.T2(lintAction, lintFunc),
	lo.T2(buildAction, buildFunc),
}

var buildProject = func(cmd *cobra.Command, args []string) {
	uArgs := lo.Uniq(args)
	attended := lo.Filter(buildActions, func(item lo.Tuple2[string, buildCmdFunc], index int) bool {
		return lo.Contains(uArgs, item.A)
	})
	// Check if the folder exists
	os.Mkdir(targetFolder, 0755)
	for _, a := range attended {
		msg := fmt.Sprintf("Start %s project", a.A)
		fmt.Printf("%-20s ...... \n", msg)
		if err := a.B(cmd); err != nil {
			internal.Red.Printf("Failed to %s project %v \n", a.A, err.Error())
			break
		}
	}
}
