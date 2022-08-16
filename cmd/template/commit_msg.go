//go:build gbt

package main

import "github.com/kcmvp/gbt/builder"

func main() {
	//input, _ := os.ReadFile(os.Args[1])
	builder.NewBuilder().Run(builder.Lint, builder.Test)
}
