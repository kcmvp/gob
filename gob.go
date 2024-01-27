/*
Copyright © 2023 kcheng.mvp@gmail.com
*/
package main

import (
	"github.com/kcmvp/gob/cmd"
	"os" //nolint
)

func main() {
	if cmd.Execute() != nil {
		os.Exit(1)
	}
	os.Exit(0)
}
