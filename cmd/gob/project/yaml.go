package project

import (
	"github.com/google/yamlfmt"
	"github.com/google/yamlfmt/formatters/basic"
	"github.com/samber/mo"
	"os"
)

func prettyYaml(fileName string) {
	register := yamlfmt.NewFormatterRegistry(&basic.BasicFormatterFactory{})
	if factory := mo.TupleToResult(register.GetDefaultFactory()); factory.IsOk() {
		if formatter := mo.TupleToResult(factory.MustGet().NewFormatter(nil)); formatter.IsOk() {
			od, _ := os.ReadFile(fileName)
			fd, _ := formatter.MustGet().Format(od)
			os.WriteFile(fileName, fd, os.ModePerm)
		}
	}
}
