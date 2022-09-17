//go:build gob

package main

import (
	"os"

	"github.com/kcmvp/gob/boot"
	"github.com/kcmvp/gob/builder"
)

func main() {
	if err := boot.NewSession().Run(builder.NewBuilder()); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}
