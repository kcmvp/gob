//go:build gbt

package main

import (
	"github.com/kcmvp/gob/builder"
	"path/filepath"
	"runtime"
)

func main() {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("No caller information")
	}
	root := filepath.Dir(filepath.Dir(filename))
	builder.NewBuilder().Run(builder.Clean, builder.Lint, builder.Test, builder.Build)
}
