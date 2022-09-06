package boot

import (
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
	if f, err = os.OpenFile(targetName, flag, os.ModePerm); err == nil {
		defer f.Close()
		if t, err = template.New(targetName).Parse(tmpl); err != nil {
			err = fmt.Errorf("failed to parse template: %w", err)
		} else {
			if err = t.Execute(f, data); err != nil {
				err = fmt.Errorf("failed to create file %v: %w", filepath.Base(targetName), err)
			}
		}
	} else if errors.Is(err, os.ErrExist) {
		log.Println(color.YellowString("%s exists", filepath.Base(targetName)))
		return nil
	}
	return err
}
