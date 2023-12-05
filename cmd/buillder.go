package cmd

import (
	"bufio"
	"fmt"
	"github.com/fatih/color"
	"github.com/kcmvp/gob/internal"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	CleanCmd           = "clean"
	TestCmd            = "test"
	LintCmd            = "lint"
	BuildCmd           = "build"
	TargetFolder       = "target"
	CleanCacheFlag     = "cache"
	CleanTestCacheFlag = "testcache"
	CleanModCacheFlag  = "modcache"
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
	fmt.Println("Clean build folder")
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

var builderFuncs = []lo.Tuple2[string, func(cmd *cobra.Command, args []string) error]{
	lo.T2(CleanCmd, func(cmd *cobra.Command, args []string) error {
		if err := cleanTarget(); err != nil {
			return err
		}
		if err := cleanBuiltIn(cmd); err != nil {
			return err
		}
		return nil
	}),
	lo.T2(TestCmd, func(cmd *cobra.Command, args []string) error {
		fmt.Println(TestCmd)
		return nil
	}),
	lo.T2(LintCmd, func(cmd *cobra.Command, args []string) error {
		fmt.Println(LintCmd)
		return nil
	}),
	lo.T2(BuildCmd, func(cmd *cobra.Command, args []string) error {
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
				// build
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

var build = func(cmd *cobra.Command, args []string) error {
	var err error
	if LintCmd == cmd.Name() {
		if t, ok := lo.Find(builderFuncs, func(item lo.Tuple2[string, func(cmd *cobra.Command, args []string) error]) bool {
			return LintCmd == item.A
		}); ok {
			fmt.Println(blueMsg.Sprintf("Start %s", cmd.Name()))
			err = t.B(cmd, args)
		}
	} else {
		fns := lo.DropRightWhile(builderFuncs, func(t2 lo.Tuple2[string, func(cmd *cobra.Command, args []string) error]) bool {
			return t2.A != cmd.Name()
		})
		// pass down flags
		for _, fn := range fns {
			fmt.Println(blueMsg.Sprintf("Start %s", fn.A))
			if err = fn.B(cmd, args); err != nil {
				break
			}
		}
	}
	if err != nil {
		return fmt.Errorf(internal.Red.Sprintf(err.Error()))
	}
	return nil
}

// cache the same as 'go clean -cache'
var cache bool

// testCache the same as `go clean -testcache'
var testCache bool

// modCache the same as 'go clean -modcache'
var modCache bool
var cleanCmd = &cobra.Command{
	Use:   CleanCmd,
	Short: "Clean target folder and build caches",
	Long:  `Clean target folder and build caches`,
	RunE:  build,
}

// report generate test or lint report
var report bool
var testCmd = &cobra.Command{
	Use:   TestCmd,
	Short: "Run tests of the project",
	Long:  `Run tests of the project and test report will be generated at ${root}/target`,
	RunE:  build,
}

var lintCmd = &cobra.Command{
	Use:   LintCmd,
	Short: "Build all main packages in project",
	Long:  `Build all main packages in project`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		// a warning message will be print when golangci-lint is not found and lintCmd is called separately
		// and gob will try to install golangci-lint
		// @todo
		// a warning message will be print when golangci-lint is not found and lintCmd is called separately
		// @todo
		return nil
	},
	RunE: build,
}
var buildCmd = &cobra.Command{
	Use:   BuildCmd,
	Short: "Build all main packages in the project",
	Long:  `Build all main packages in the project`,
	//RunE:  build,
	RunE: func(cmd *cobra.Command, args []string) error {
		v, err := cmd.Flags().GetBool(CleanCacheFlag)
		println(v)
		return err
	},
}

func init() {
	viper.SetDefault("ContentDir", "content")
	buildCmd.Flags().BoolVar(&cache, CleanCacheFlag, true, "to remove the entire go build cache")
	buildCmd.Flags().BoolVar(&testCache, CleanTestCacheFlag, true, "to expire all test results in the go build cache")
	buildCmd.Flags().BoolVar(&modCache, CleanModCacheFlag, true, "to remove the entire module download cache")
	buildCmd.Flags().BoolVar(&report, "report", true, "generate build report")
	lintCmd.Flags().BoolVar(&report, "report", false, "generate lint report")
	testCmd.Flags().BoolVar(&report, "report", false, "generate test report")
	cleanCmd.Flags().BoolVar(&cache, CleanCacheFlag, false, "to remove the entire go build cache")
	cleanCmd.Flags().BoolVar(&testCache, CleanTestCacheFlag, false, "to expire all test results in the go build cache")
	cleanCmd.Flags().BoolVar(&modCache, CleanModCacheFlag, false, "to remove the entire module download cache")
	rootCmd.AddCommand(cleanCmd)
	rootCmd.AddCommand(testCmd)
	rootCmd.AddCommand(lintCmd)
	rootCmd.AddCommand(buildCmd)
}
