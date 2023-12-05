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

var blueMsg = color.New(color.FgCyan)

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

func cleanTarget() error {
	fmt.Println("Clean action folder")
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
	return nil
}
func cleanBuiltIn(cmd *cobra.Command) error {
	args := []string{"clean"}
	if ok, _ := cmd.Flags().GetBool(CleanCacheFlag); ok {
		fmt.Println("clean is true")
		args = append(args, fmt.Sprintf("-%s", CleanCacheFlag))
	}
	if ok, _ := cmd.Flags().GetBool(CleanTestCacheFlag); ok {
		fmt.Println("clean test is true")
		args = append(args, fmt.Sprintf("-%s", CleanTestCacheFlag))
	}
	if ok, _ := cmd.Flags().GetBool(CleanModCacheFlag); ok {
		fmt.Println("clean mode is true")
		args = append(args, fmt.Sprintf("-%s", CleanModCacheFlag))
	}
	_, err := exec.Command("go", args...).CombinedOutput()
	return err
}

var actions = []lo.Tuple3[string, int, func(cmd *cobra.Command) error]{
	lo.T3(cleanAction, 0, func(cmd *cobra.Command) error {
		if err := cleanTarget(); err != nil {
			return err
		}
		if err := cleanBuiltIn(cmd); err != nil {
			return err
		}
		return nil
	}),
	lo.T3(testAction, 1, func(cmd *cobra.Command) error {
		fmt.Println(testAction)
		return nil
	}),
	lo.T3(lintAction, 2, func(cmd *cobra.Command) error {
		fmt.Println(lintAction)
		return nil
	}),
	lo.T3(buildAction, 100, func(cmd *cobra.Command) error {
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
				os.MkdirAll(filepath.Join(internal.CurProject().Root(), TargetFolder), 0755)
				if _, err := exec.Command("go", "build", "-o", output, mf).CombinedOutput(); err != nil { //nolint
					return err
				} else {
					fmt.Printf("Build %s to %s successfully\n", mf, output)
					bm[binary] = output
				}
			} else {
				fmt.Println(internal.Yellow.Sprintf("Can not find main function in package %s", dir))
			}
		}
		return nil
	}),
}

var execBuild = func(cmd *cobra.Command, args []string) error {
	var err error
	maxCmd := lo.MaxBy(args, func(a string, b string) bool {
		ta, _ := lo.Find(actions, func(item lo.Tuple3[string, int, func(cmd *cobra.Command) error]) bool {
			return item.A == a
		})
		tb, _ := lo.Find(actions, func(item lo.Tuple3[string, int, func(cmd *cobra.Command) error]) bool {
			return item.A == b
		})
		return ta.B > tb.B
	})
	if len(args) == 1 && maxCmd == lintAction {
		if t3, ok := lo.Find(actions, func(item lo.Tuple3[string, int, func(cmd *cobra.Command) error]) bool {
			return item.A == lintAction
		}); ok {
			fmt.Println(blueMsg.Sprintf("Start %s", lintAction))
			err = t3.C(cmd)
		}
	} else {
		t3s := lo.DropRightWhile(actions, func(t3 lo.Tuple3[string, int, func(cmd *cobra.Command) error]) bool {
			return t3.A != maxCmd
		})
		// pass down flags
		for _, t3 := range t3s {
			fmt.Println(blueMsg.Sprintf("Start %s", t3.A))
			if err = t3.C(cmd); err != nil {
				break
			}
		}
	}
	if err != nil {
		return fmt.Errorf(internal.Red.Sprintf(err.Error()))
	}
	return nil
}
