/*
Copyright © 2023 kcheng.mvp@gmail.com
*/
package main

import (
	"github.com/kcmvp/gob/cmd/gbc/command"
	"os" //nolint
)

func main() {
	if err := command.Execute(); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}
