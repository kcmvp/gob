/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"github.com/kcmvp/gob/cmd"
	"github.com/kcmvp/gob/scaffolds"
)

func init() {
	scaffolds.NewProject()
}

func main() {
	cmd.Execute()
}
