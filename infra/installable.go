package infra

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/fatih/color"
)

const LatestVer = "latest"

type VerFunc func(cmd string) string

type Installable interface {
	Install(ver string) (string, error)
	Cmd() string
	Installed() []string
	Version(cmd string) string
}

type defaultIns struct {
	module, cmd string
	verFunc     VerFunc
}

func (i *defaultIns) Version(cmd string) string {
	return i.verFunc(cmd)
}

func (i *defaultIns) Cmd() string {
	return i.cmd
}

var _ Installable = (*defaultIns)(nil)

func NewInstallable(module, cmd string, verFunc VerFunc) Installable {
	return &defaultIns{
		module,
		cmd,
		verFunc,
	}
}

func (i *defaultIns) Installed() []string {
	// Executables are installed in the directory named by the GOBIN environment
	// variable, which defaults to $GOPATH/bin or $HOME/go/bin if the GOPATH
	// environment variable is not set.
	vm := map[string]byte{}
	h, _ := os.UserHomeDir()
	goBin := filepath.Join(h, "go", "bin")
	filepath.Walk(goBin, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasPrefix(info.Name(), i.cmd) {
			v := i.verFunc(path)
			vm[v] = 1
			i.tagVersion(path, v)
		}
		return err
	})
	var vs []string
	for s := range vm {
		vs = append(vs, s)
	}
	if len(vs) > 0 {
		var desc []string
		for idx, v := range vs {
			desc = append(desc, fmt.Sprintf("%s):%s", strconv.Itoa(idx+1), v))
		}
		log.Printf("installed versions of %s: %s \n", i.cmd, strings.Join(desc, ", "))
	}
	return vs
}

func (i *defaultIns) Install(ver string) (string, error) {
	var err error
	var out []byte
	ivs := i.Installed()
	for _, v := range ivs {
		if v == ver {
			return ver, nil
		}
	}
	if len(ivs) > 0 && LatestVer == ver {
		fmt.Println("please select version number:")
		completer := func(d prompt.Document) []prompt.Suggest {
			var s []prompt.Suggest
			for idx, v := range ivs {
				s = append(s, prompt.Suggest{Text: strconv.Itoa(idx + 1), Description: fmt.Sprintf("using %s", v)})
			}
			return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)
		}
		v := prompt.Input(">> ", completer)
		if idx, err := strconv.Atoi(v); err == nil && idx >= 1 && idx <= len(ivs) {
			v = ivs[idx-1]
			fmt.Printf("using existing %s\n", v)
			return v, nil
		}
	}

	vm := fmt.Sprintf("%s@%s", i.module, ver)
	log.Printf("installing %s ...\n", vm)
	cmd := exec.Command("go", "install", vm)
	out, err = cmd.CombinedOutput()
	if err != nil {
		log.Println(string(out))
		log.Println(color.RedString("failed to install %s", vm))
		log.Println(color.RedString("you can manually install it by 'go install %s'", vm))
	} else {
		ver = i.Version(i.Cmd())
		vm = fmt.Sprintf("%s@%s", i.module, ver)
		log.Printf("%s is installed successfully\n", vm)
		i.tagVersion(cmd.Path, ver)
	}

	return ver, err
}

func (i *defaultIns) tagVersion(file, ver string) {
	base := filepath.Base(file)
	if strings.HasPrefix(base, i.Cmd()) && strings.Contains(base, ver) {
		return
	}
	target := fmt.Sprintf("%s-%s", i.Cmd(), ver)
	if strings.HasSuffix(base, ".exe") {
		target = fmt.Sprintf("%s.exe", target)
	}
	if t, err := os.OpenFile(filepath.Join(filepath.Dir(file), target), os.O_RDWR|os.O_CREATE|os.O_EXCL, os.ModePerm); err == nil {
		if s, err := os.Open(file); err == nil {
			if _, err = io.Copy(t, s); err != nil {
				log.Fatalln(color.RedString("failed to tag %s as %s", filepath.Base(file), target))
			}
		}
	} else if !errors.Is(err, os.ErrExist) {
		log.Fatalln(color.RedString("failed to tag %s as %s", filepath.Base(file), target))
	}
}
