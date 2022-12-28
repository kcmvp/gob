//go:build gob

package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/kcmvp/gob/boot"
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	line, _ := reader.ReadString('\n')
	refs := strings.Fields(line)
	// do nothing for push delete, merge
	if len(refs) > 0 && strings.Contains(refs[0], "delete") {
		os.Exit(0)
	}
	fmt.Println(refs)

	pwd, _ := os.Getwd()
	if err := boot.NewSession().Run(boot.NewProject(pwd)); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}
