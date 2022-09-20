package boot

import (
	"go/build"
	"log"
	"os"
	"runtime"
	"strings"
)

type Inspector[T comparable] func(frame string) T

func Inspect[T comparable](inspector Inspector[T]) T {
	goRoot := runtime.GOROOT()
	log.Printf("GOROOT: %s\n", goRoot)
	goPath := os.Getenv("GOPATH")
	if goPath == "" {
		goPath = build.Default.GOPATH
	}
	log.Printf("GOPATH: %s\n", goPath)
	runtimePaths := []string{goRoot, goPath}

	pc := make([]uintptr, 15)   //nolint
	n := runtime.Callers(1, pc) //nolint
	frames := runtime.CallersFrames(pc[:n])
	var frame runtime.Frame
	more := true
	var zero T
	for more {
		frame, more = frames.Next()
		fromGoRuntime := false
		for _, path := range runtimePaths {
			if strings.HasPrefix(frame.File, path) {
				fromGoRuntime = true
				break
			}
		}
		if !fromGoRuntime {
			if t := inspector(frame.File); t != zero {
				return t
			}
		}
	}
	return zero
}
