//go:build gbt

package main

import (
	"github.com/fatih/color"
	"github.com/kcmvp/gbt/builder"
	"log"
	"os"
)

func main() {
	input, _ := os.ReadFile(os.Args[1])
	project := builder.NewProject(builder.DefaultHookCfg()).Clean().Test().Scan(string(input))
	if project.Quality().LinterIssues.Files > 0 {
		log.Fatalln(color.RedString("failed to commit the code"))
	}
	os.Exit(0)
}
