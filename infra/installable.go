package infra

import (
	"fmt"
	"github.com/fatih/color"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

type Version func() string

type Installable interface {
	Install(ver string, currVer Version) (string, error)
	Cmd() string
	Versions() []string
}

type defaultIns struct {
	module, cmd string
}

func (i *defaultIns) Cmd() string {
	return i.cmd
}

var _ Installable = (*defaultIns)(nil)

func NewInstallable(module, cmd string) Installable {
	return &defaultIns{
		module: module,
		cmd:    cmd,
	}
}

func (i *defaultIns) Versions() []string {
	// Executables are installed in the directory named by the GOBIN environment
	//variable, which defaults to $GOPATH/bin or $HOME/go/bin if the GOPATH
	//environment variable is not set.
	var vs []string
	h, _ := os.UserHomeDir()
	goBin := filepath.Join(h, "go", "bin")
	sep := fmt.Sprintf("%s-", i.cmd)
	filepath.WalkDir(goBin, func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() {
			//@todo fix windows naming issue
			items := strings.Split(d.Name(), sep)
			if len(items) == 1 && d.Name() != items[0] {
				vs = append(vs, items[0])
			}
		}
		return nil
	})
	return vs
}

func (i *defaultIns) Install(ver string, verFun Version) (string, error) {
	var err error
	var out []byte
	installed := false

	vs := append(i.Versions(), verFun())
	for _, v := range vs {
		if installed = ver == v; installed {
			log.Printf("found existing version of %s\n", i.cmd)
			break
		}
	}
	if !installed {
		vm := fmt.Sprintf("%s@%s", i.module, ver)
		log.Printf("installing %s ...\n", vm)
		cmd := exec.Command("go", "install", vm)
		out, err = cmd.CombinedOutput()
		if err != nil {
			log.Println(string(out))
			log.Println(color.RedString("failed to install %s", vm))
			log.Println(color.RedString("you can manually install it by 'go install %s'", vm))
		} else {
			ver = verFun()
			vm = fmt.Sprintf("%s@%s", i.module, ver)
			log.Printf("%s is installed successfully\n", vm)
			versionedCmd := fmt.Sprintf("%s-%s", i.cmd, ver)
			if runtime.GOOS == "windows" {
				versionedCmd = fmt.Sprintf("%s.exe", versionedCmd)
			}
			tagCmd(cmd.Path, filepath.Join(filepath.Dir(cmd.Path), versionedCmd))
		}
	}
	return ver, err
}

func tagCmd(src, target string) error {
	//@todo, backup
	return nil
}
