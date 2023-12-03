/*
Copyright © 2023 kcheng.mvp@gmail.com
*/
package main

import (
	"github.com/kcmvp/gob/cmd"
	"os"
)

func main() {
	os.Exit(cmd.Execute(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}
