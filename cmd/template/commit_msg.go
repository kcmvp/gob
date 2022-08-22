//go:build gos

package main

import (
	"github.com/kcmvp/gos/builder"
	"path/filepath"
	"runtime"
)

func main() {
	//input, _ := os.ReadFile(os.Args[1])

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("No caller information")
	}
	root := filepath.Dir(filepath.Dir(filename))
	builder.NewBuilder(root).Run(builder.Lint, builder.Test)
}
