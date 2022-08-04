////go:build gbt

package main

import (
	"github.com/kcmvp/gbt/builder"
)

func main() {
	// default coverage min:0.35, max: 0.85; by default for each push then coverage can not degrease, maxCoverage means
	// if current coverage is bigger or equals maxCoverage then there is no such check
	min, max := 0.35, 0.85
	project := builder.NewProject(min, max)
	project.Clean().Test().Scan().Report().Build()
}
