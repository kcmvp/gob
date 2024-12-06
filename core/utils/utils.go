package utils

import (
	"github.com/samber/lo"
	"github.com/samber/mo"
	"regexp"
	"runtime"
	"strings"
)

// CleanStr Function to remove non-printable characters
func CleanStr(str string) string {
	cleanStr := func(r rune) rune {
		if r >= 32 && r != 127 {
			return r
		}
		return -1
	}
	return strings.Map(cleanStr, str)
}

// TestEnv returns full unique of the method name together a bool value
// true indicates the caller is from  _test.go. As init() is executed before any
// other method, so call this method in init() would not return correct result.
func TestEnv() mo.Option[string] {
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
	return mo.TupleToOption[string](fqn, test)
}
