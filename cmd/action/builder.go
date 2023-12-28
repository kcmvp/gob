package action

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/fatih/color"
	"github.com/kcmvp/gob/internal"
	"github.com/spf13/cobra"
)

var (
	CleanCache     bool
	CleanTestCache bool
	CleanModCache  bool
)

const (
	CleanCacheFlag     = "cache"
	CleanTestCacheFlag = "testcache"
	CleanModCacheFlag  = "modcache"
)

var builtinActions = []CmdAction{
	{A: "build", B: buildCommand},
	{A: "clean", B: cleanCommand},
	{A: "test", B: testCommand},
}

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

var buildCommand = func(_ *cobra.Command, args ...string) error {
	dirs, err := internal.FindGoFilesByPkg("main")
	if err != nil {
		return err
	}
	bm := map[string]string{}
	for _, dir := range dirs {
		mainFile, err := findMain(dir)
		if err != nil {
			return err
		}
		if len(mainFile) > 0 {
			// action
			binary := strings.TrimSuffix(filepath.Base(mainFile), ".go")
			if f, exists := bm[binary]; exists {
				return fmt.Errorf("file %s has already built as %s, please rename %s", f, binary, mainFile)
			}
			output := filepath.Join(internal.CurProject().Target(), binary)
			versionFlag := fmt.Sprintf("-X 'main.buildVersion=%s'", internal.Version())
			// try to build the binary with version first
			if _, err := exec.Command("go", "build", "-ldflags", versionFlag, "-o", output, mainFile).CombinedOutput(); err != nil { //nolint
				color.Yellow("no version variable 'buildVersion' defined in the main package")
				if _, err := exec.Command("go", "build", "-o", output, mainFile).CombinedOutput(); err != nil {
					return err
				}
			}
			fmt.Printf("Build %s to %s successfully\n", mainFile, output)
			bm[binary] = output
		} else {
			color.Yellow("Can not find main function in package %s", dir)
		}
	}
	//
	return nil
}

var cleanCommand = func(cmd *cobra.Command, _ ...string) error {
	// clean target folder
	os.RemoveAll(internal.CurProject().Target())
	os.Mkdir(internal.CurProject().Target(), os.ModePerm) //nolint errcheck
	fmt.Println("Clean target folder successfully !")
	// clean cache
	args := []string{"clean"}
	if CleanCache {
		args = append(args, fmt.Sprintf("--%s", CleanCacheFlag))
	}
	if CleanTestCache {
		args = append(args, fmt.Sprintf("--%s", CleanTestCacheFlag))
	}
	if CleanModCache {
		args = append(args, fmt.Sprintf("--%s", CleanModCacheFlag))
	}
	_, err := exec.Command("go", args...).CombinedOutput()
	if len(args) > 1 && err == nil {
		fmt.Println("Clean cache successfully !")
	}
	return nil
}

var testCommand = func(_ *cobra.Command, args ...string) error {
	coverProfile := fmt.Sprintf("-coverprofile=%s/cover.out", internal.CurProject().Target())
	testCmd := exec.Command("go", []string{"test", "-v", coverProfile, "./..."}...) //nolint
	err := StreamExtCmdOutput(testCmd, fmt.Sprintf("%s/test.log", internal.CurProject().Target()), "FAIL:")
	if err != nil {
		return err
	}
	_, err = exec.Command("go", []string{"tool", "cover", fmt.Sprintf("-html=%s/cover.out", internal.CurProject().Target()), fmt.Sprintf("-o=%s/cover.html", internal.CurProject().Target())}...).CombinedOutput() //nolint
	if err == nil {
		color.Green("Test report is generated at %s/test.log \n", internal.CurProject().Target())
		color.Green("Coverage report is generated at %s/cover.html \n", internal.CurProject().Target())
	}
	return nil
}
