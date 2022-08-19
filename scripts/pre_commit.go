//go:build gos

package main

import "github.com/kcmvp/gos/builder"

func main() {
	//input, _ := os.ReadFile(os.Args[1])
	builder.NewBuilder().Run()
}
