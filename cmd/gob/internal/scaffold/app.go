package scaffold

import (
	_ "embed"
	"github.com/samber/lo"
)

//go:embed resources/modules.json
var data []byte

type Module lo.Tuple2[string, string]

func Supported() []Module {
	panic("")
}

func Installed() []Module {
	panic("")
}

func Install(module Module) error {
	panic("")
}
