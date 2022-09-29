//go:build gob

package main

import (
	"github.com/kcmvp/gob/scaffolds"
	"os"

	"github.com/kcmvp/gob/boot"
)

func main() {
	// input, _ := os.ReadFile(os.Args[1])

	if err := boot.NewSession().Run(scaffolds.NewProject()); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}
