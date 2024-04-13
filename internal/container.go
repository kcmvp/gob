package internal

import (
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/kcmvp/gob/utils"
	"github.com/samber/do/v2"
)

var (
	Container *do.RootScope
	RootDir   string
)

func init() {
	Container = do.NewWithOpts(&do.InjectorOpts{
		HookAfterRegistration: func(scope *do.Scope, serviceName string) {
			fmt.Printf("scope is %s, name is %s \n", scope.Name(), serviceName)
			//@todo, parse the mapping once
		},
		Logf: func(format string, args ...any) {
			log.Printf(format, args...)
		},
	})
	if output, err := exec.Command("go", "list", "-f", "{{.Root}}").CombinedOutput(); err == nil {
		RootDir = utils.CleanStr(string(output))
	} else {
		RootDir, _ = os.Executable()
	}
}
