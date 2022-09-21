//go:build gob

package main

import (
	"bufio"
	"fmt"
	"github.com/kcmvp/gob/scaffolds"
	"os"
	"strings"

	"github.com/kcmvp/gob/boot"
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

	if err := boot.NewSession().Run(scaffolds.NewProject()); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}
