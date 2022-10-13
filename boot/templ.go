package boot

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"text/template"

	"github.com/fatih/color"
)

func GenerateFile(tmpl string, targetName string, data interface{}, trunk bool) error {
	flag := os.O_RDWR | os.O_CREATE | os.O_EXCL //nolint:nosnakecase
	if trunk {
		flag = os.O_RDWR | os.O_CREATE | os.O_TRUNC //nolint:nosnakecase
	}
	var err error
	var f *os.File
	var t *template.Template
	if f, err = os.OpenFile(targetName, flag, os.ModePerm); err == nil { //nolint:nestif
		defer f.Close()
		if t, err = template.New(targetName).Parse(tmpl); err != nil {
			err = fmt.Errorf("failed to parse template: %w", err)
		} else {
			if err = t.Execute(f, data); err != nil {
				err = fmt.Errorf("failed to create file %v: %w", filepath.Base(targetName), err)
			}
		}
	} else if !trunk && errors.Is(err, os.ErrExist) {
		// it's normal get os.ErrExist when don't trunk existing file
		log.Println(color.YellowString("File: %s exists", filepath.Base(targetName)))
		err = nil
	}
	return err
}

func GenerateString(tmpl string, data interface{}) string {
	var buf bytes.Buffer
	if t, err := template.New("tmpl").Parse(tmpl); err != nil {
		log.Fatalln(color.YellowString("failed to parse template: %s", err.Error()))
	} else {
		if err = t.Execute(&buf, data); err != nil {
			log.Fatalln(color.YellowString("failed to create string %v: %s", data, err.Error()))
		}
	}
	return buf.String()
}
