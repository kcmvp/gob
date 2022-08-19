//go:build gbt

package main

import "github.com/kcmvp/gos/builder"

func main() {
	builder.NewBuilder().Run(builder.Clean, builder.Lint, builder.Test, builder.Build)
}
