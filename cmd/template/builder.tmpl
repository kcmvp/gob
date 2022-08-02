//go:build gbt

package main

import (
	"context"
	"flag"
	"github.com/kcmvp/gbt/builder"
)

func main() {
	// default coverage min:0.35, max: 0.85; by default for each push then coverage can not degrease, maxCoverage means
	// if current coverage is bigger or equals maxCoverage then there is no such check
	var event string
	flag.StringVar(&event, "event", "", "git event")
	flag.Parse()
	ctx := context.Background()
	ctx = context.WithValue(ctx, "event", event)
	project := builder.NewProject(0.35, 0.85).WithCtx(ctx)
	project.Clean().Test().Scan().Check().Build()
}
