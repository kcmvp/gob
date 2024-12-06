package env

import (
	"github.com/kcmvp/gob/core/utils"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
)

var (
	rootDir string
	once    sync.Once
)

const (
	DefaultCfg = "application"
)

func RootDir() string {
	if len(rootDir) == 0 {
		once.Do(func() {
			dir, _ := exec.Command("go", "list", "-m", "-f", "{{.Dir}}").CombinedOutput()
			rootDir = utils.CleanStr(string(dir))
			if len(rootDir) == 0 {
				binary, _ := os.Executable()
				rootDir = filepath.Dir(binary)
			}
		})
	}
	return rootDir
}

// WindowsEnv return true when current os is WindowsEnv
func WindowsEnv() bool {
	return runtime.GOOS == "windows"
}
