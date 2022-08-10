//go:build gbt

package main

import (
	"bufio"
	"fmt"
	"github.com/kcmvp/gbt/builder"
	"os"
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
	builder.NewProject(builder.DefaultHookCfg()).Clean().Test().Scan(refs)
}
