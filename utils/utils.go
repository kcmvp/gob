package utils

import (
	"regexp"
	"runtime"
	"strings"

	"github.com/samber/lo"
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

func TestCaller() (bool, string) {
	var test bool
	var frame runtime.Frame
	more := true
	callers := make([]uintptr, 20)
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
	return test, fqn
}
