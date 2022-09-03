package boot

import (
	"bufio"
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
	"github.com/samber/lo"
)

const LatestVer = "latest"

type Version func(cmd string) string

type Installer interface {
	Install(ver string) (string, error)
	Cmd() string
	Installed() []string
	Configured() (string, error)
	Version(cmd string) string
}

type installer struct {
	module  string
	cmd     string
	config  string
	version Version
}

func (ins *installer) Version(cmd string) string {
	return ins.version(cmd)
}

func (ins *installer) Cmd() string {
	return ins.cmd
}

var _ Installer = (*installer)(nil)

func NewInstallable(module, cmd, config string, version Version) Installer {
	return &installer{
		module,
		cmd,
		config,
		version,
	}
}

func (ins *installer) Configured() (string, error) {
	var ver string
	var err error
	f, err := os.Open(ins.config)
	defer f.Close()
	if err == nil {
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := scanner.Text()
			if items := strings.Split(line, "version:"); len(items) == 2 {
				ver = strings.TrimSpace(items[1])
				log.Printf("linter is configured as %s \n", ver)
				break
			}
		}
		if ver == "" {
			msg := color.RedString("missed version in %s", ins.config)
			log.Println(msg)
			err = fmt.Errorf(msg)
		}
	}
	return ver, err
}

func (ins *installer) Installed() []string {
	// Executables are installed in the directory named by the GOBIN environment
	// variable, which defaults to $GOPATH/bin or $HOME/go/bin if the GOPATH
	// environment variable is not set.
	vMap := map[string]byte{}
	h, _ := os.UserHomeDir()
	goBin := filepath.Join(h, "go", "bin")
	err := filepath.WalkDir(goBin, func(path string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() && strings.HasPrefix(d.Name(), ins.cmd) {
			v := ins.version(path)
			vMap[v] = 1
			ins.tagVersion(path, v)
		}
		return err
	})
	if err != nil {
		log.Fatalln(color.RedString("failed to get installed version of %s:%s", ins.cmd, err.Error()))
	}
	versions := lo.MapToSlice(vMap, func(k string, v byte) string {
		return k
	})

	desc := lo.Map(versions, func(t string, i int) string {
		return fmt.Sprintf("%d):%s", i+1, t)
	})
	if len(desc) > 0 {
		log.Printf("installed versions of %s: %s \n", ins.cmd, strings.Join(desc, ", "))
	}
	return versions
}

func (ins *installer) Install(ver string) (string, error) {
	var err error
	var out []byte
	ivs := ins.Installed()
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

	vm := fmt.Sprintf("%s@%s", ins.module, ver)
	log.Printf("installing %s ...\n", vm)
	cmd := exec.Command("go", "install", vm)
	out, err = cmd.CombinedOutput()
	if err != nil {
		log.Println(string(out))
		log.Println(color.RedString("failed to install %s", vm))
		log.Println(color.RedString("you can manually install it by 'go install %s'", vm))
	} else {
		ver = ins.Version(ins.Cmd())
		vm = fmt.Sprintf("%s@%s", ins.module, ver)
		log.Printf("%s is installed successfully\n", vm)
		ins.tagVersion(cmd.Path, ver)
	}

	return ver, err
}

func (ins *installer) tagVersion(file, ver string) {
	base := filepath.Base(file)
	if strings.HasPrefix(base, ins.Cmd()) && strings.Contains(base, ver) {
		return
	}
	target := fmt.Sprintf("%s-%s", ins.Cmd(), ver)
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
