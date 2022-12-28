//go:build gob

package main

import (
	"os"

	"github.com/kcmvp/gob/boot"
)

func main() {

	pwd, _ := os.Getwd()
	if err := boot.NewSession().Run(boot.NewProject(pwd)); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}
