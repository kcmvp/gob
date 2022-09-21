//go:build gob

package main

import (
	"github.com/fatih/color"
	"log"
	"os"

	"github.com/kcmvp/gob/boot"
	"github.com/kcmvp/gob/builder"
)

func main() {
	// input, _ := os.ReadFile(os.Args[1])

	if err := boot.NewSession().Run(builder.NewProject()); err != nil {
		log.Println(color.RedString(err.Error()))
		os.Exit(1)
	}
	os.Exit(0)
}
