/*
Copyright © 2023 kcheng.mvp@gmail.com
*/
package main

import (
	"github.com/kcmvp/gob/gbc/cmd"
	"os" //nolint
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}
