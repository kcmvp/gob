//go:build gob

package main

import (
	"github.com/kcmvp/gob/boot"
	"github.com/kcmvp/gob/builder"
	"os"
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
	boot.Run(builder.NewBuilder(root))
	os.Exit(0)
}
