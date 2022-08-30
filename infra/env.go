package infra

import (
	"embed"
	"errors"
	"log"
	"path/filepath"

	"github.com/fatih/color"
)

var errEnv = errors.New("environment error")

//go:embed template/*.tmpl
var templateDir embed.FS

func SetupBuilder(dir string) error {
	var err error
	var tf []byte
	tf, err = templateDir.ReadFile(filepath.Join("template", "builder.tmpl"))
	if err != nil {
		return err
	}
	return GenerateFile(string(tf), filepath.Join(dir, "builder.go"), nil, false)
}

func CheckError(err error, msg string) {
	if err != nil {
		log.Fatalln(color.RedString("%s: %s", msg, err.Error()))
	}
}
