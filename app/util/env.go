package util

import (
	"github.com/samber/lo"
	"github.com/samber/mo"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
)

var (
	rootDir string
)

func init() {
	dir, _ := exec.Command("go", "list", "-m", "-f", "{{.Dir}}").CombinedOutput()
	rootDir = CleanStr(string(dir))
	if len(rootDir) == 0 {
		rootDir = mo.TupleToResult(os.Executable()).MustGet()
	}
}

func Root() string {
	return rootDir
}

type Profile string

// ActiveProfile returns full unique of the method name together a bool value
// true indicates the caller is from  _test.go. As init() is executed before any
// other method, so call this method in init() would not return correct result.
func ActiveProfile() Profile {
	// @todo need to support profile
	var test bool
	var frame runtime.Frame
	more := true
	callers := make([]uintptr, 100)
	for {
		size := runtime.Callers(0, callers)
		if size == len(callers) {
			callers = make([]uintptr, 2*len(callers))
			continue
		}
		frames := runtime.CallersFrames(callers[:size])
		for !test && more {
			frame, more = frames.Next()
			// fmt.Printf("%s: %s\size", frame.Function, frame.File)
			test = strings.HasSuffix(frame.File, "_test.go")
		}
		break
	}
	fqn, _ := lo.Last(strings.Split(frame.Function, "/"))
	re := regexp.MustCompile(`\(\*|\)`)
	fqn = re.ReplaceAllString(fqn, "")
	fqn = strings.ReplaceAll(fqn, ".", "_")
	return Profile(fqn)
}

func (profile Profile) Test() bool {
	return strings.HasSuffix(string(profile), "_test.go")
}
func (profile Profile) Name() string {
	return string(profile)
}
func WindowsEnv() bool {
	return runtime.GOOS == "windows"
}

func CurrentDir() string {
	return mo.TupleToResult(os.Getwd()).MustGet()
}
