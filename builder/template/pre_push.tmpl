//go:build gob

package main

import (
	"bufio"
	"fmt"
	"github.com/kcmvp/gob/builder"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	line, _ := reader.ReadString('\n')
	refs := strings.Fields(line)
	// do nothing for push delete, merge
	if strings.Contains(refs[0], "delete") {
		os.Exit(0)
	}
	fmt.Println(refs)

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("No caller information")
	}
	root := filepath.Dir(filepath.Dir(filename))

	builder.NewBuilder(root).Run()
	os.Exit(0)
}
