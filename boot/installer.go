package boot

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/samber/lo"
)

const LatestVer = "latest"

type Version func(cmd string) (string, string)

type Installer interface {
	Install(ver string) (string, error)
	Cmd() string
	Versions() []string
	Version(cmd string) (string, string)
	CfgVerKey() string
	Format(ver string) string
}

type installer struct {
	module  string
	cmd     string
	config  string
	version Version
}

func (ins *installer) CfgVerKey() string {
	return fmt.Sprintf("%s.%s", CfgPrefix, ins.Cmd())
}

func (ins *installer) Format(ver string) string {
	return strings.ReplaceAll(ver, ".", "-")
}

func (ins *installer) Version(cmd string) (string, string) {
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

func (ins *installer) Versions() []string {
	vMap := map[string]byte{}
	ver, file := ins.version(ins.Cmd())
	if ver == "" {
		return []string{}
	}
	err := filepath.WalkDir(filepath.Dir(file), func(path string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() && strings.HasPrefix(d.Name(), ins.cmd) {
			v, _ := ins.version(path)
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
		log.Printf("Found following versions of %s: %s \n", ins.cmd, strings.Join(desc, ", "))
	}
	return versions
}

func (ins *installer) Install(ver string) (string, error) {
	var err error
	var out []byte
	ivs := ins.Versions()
	for _, v := range ivs {
		if v == ver {
			return ver, nil
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
		var file string
		ver, file = ins.Version(ins.Cmd())
		vm = fmt.Sprintf("%s@%s", ins.module, ver)
		log.Printf("%s is installed successfully\n", vm)
		ins.tagVersion(file, ver)
	}

	return ver, err
}

func (ins *installer) tagVersion(file, ver string) {
	fv := ins.Format(ver)
	base := filepath.Base(file)
	if strings.HasPrefix(base, ins.Cmd()) && strings.Contains(base, fv) {
		return
	}
	if _, err := os.Readlink(file); err == nil {
		return
	}
	target := fmt.Sprintf("%s-%s", ins.Cmd(), fv)
	if strings.HasSuffix(base, ".exe") {
		target = fmt.Sprintf("%s.exe", target)
	}
	target = filepath.Join(filepath.Dir(file), target)
	if err := os.Rename(file, target); err != nil {
		log.Println(color.RedString("Failed tag the file %s to %s", file, target))
	} else {
		if err = os.Symlink(target, file); err != nil {
			log.Println(color.RedString("Failed to create soft link of %s", target))
		}
	}
}
