package infra

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"log"
	"path/filepath"

	"github.com/fatih/color"
)

var errEnv = errors.New("environment error")

//go:embed template/*.tmpl
var templateDir embed.FS

type EnvCtxKey string

const ProjectRootDir EnvCtxKey = "_projectRootDir"

func root(ctx context.Context) (string, error) {
	if v, ok := ctx.Value(ProjectRootDir).(string); ok {
		return v, nil
	}
	return "", fmt.Errorf("%w: %s", errEnv, "can't find project root dir")
}

func SetupBuilder(dir string) {
	var err error
	var tf []byte
	tf, err = templateDir.ReadFile(filepath.Join("template", "builder.tmpl"))
	if err == nil {
		err = GenerateFile(string(tf), filepath.Join(dir, "builder.go"), nil, false)
	}
	CheckError(err, "Failed to create builder.go")
}

func CheckError(err error, msg string) {
	if err != nil {
		log.Fatalln(color.RedString("%s: %s", msg, err.Error()))
	}
}
