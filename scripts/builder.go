//go:build gbt

package main

import (
	"github.com/kcmvp/gob/boot"
	"github.com/kcmvp/gob/builder"
)

func main() {
	boot.Run(builder.NewBuilder(), boot.Clean, boot.Build)
}
