package infra

import (
	"errors"
	"github.com/fatih/color"
	"log"
	"os"
	"path/filepath"
	"text/template"
)

func GenerateFile(content string, targetName string, data interface{}, trunk bool) error {
	dir := filepath.Dir(targetName)
	os.MkdirAll(dir, os.ModePerm)
	flag := os.O_RDWR | os.O_CREATE | os.O_EXCL
	if trunk {
		flag = os.O_RDWR | os.O_CREATE | os.O_TRUNC
	}
	var err error
	var f *os.File
	var t *template.Template

	if f, err = os.OpenFile(targetName, flag, os.ModePerm); err == nil {
		defer f.Close()
		if t, err = template.New(targetName).Parse(content); err != nil {
			log.Println(color.RedString("Failed to parse template, %+v", err))
		} else {
			if err = t.Execute(f, data); err != nil {
				log.Println(color.RedString("Failed to create file %v, %+v\n", filepath.Base(targetName), err))
			}
		}
	} else {
		if !errors.Is(err, os.ErrExist) {
			log.Println(color.RedString("failed to generate file %s, %v\n", filepath.Base(targetName), err))
		}
	}
	return err
}
