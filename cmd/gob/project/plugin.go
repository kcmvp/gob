package project

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/kcmvp/gob/core/env"
	"github.com/samber/lo" //nolint
	"github.com/samber/mo"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

const urlPattern = `^[^@]+@[^@]+$`

type Plugin struct {
	Name        string   `json:"name"`
	Args        []string `json:"args" mapstructure:"args"`
	Url         string   `json:"url" mapstructure:"url"` //nolint
	Module      string   `json:"module"`
	Shell       string   `json:"shell"`
	Config      string   `json:"config"`
	Description string   `json:"description" mapstructure:"description"`
}

func (plugin *Plugin) UnmarshalJSON(data []byte) error {
	type Embedded Plugin
	wrapper := &struct{ *Embedded }{(*Embedded)(plugin)}
	if err := json.Unmarshal(data, &wrapper); err != nil {
		return err
	}
	if len(plugin.Url) > 0 {
		module := lo.If(len(plugin.Module) > 0, plugin.Module).Else(plugin.Url)
		rs := mo.TupleToResult(exec.Command("go", append([]string{"list", "-m"}, fmt.Sprintf("%s@latest", module))...).CombinedOutput())
		if rs.IsError() {
			return rs.Error()
		}
		plugin.Url = fmt.Sprintf("%s@%s", plugin.Url, strings.Fields(string(rs.MustGet()))[1])
	}
	return nil
}

func (plugin Plugin) version() mo.Option[string] {
	if strings.Count(plugin.Url, "@") != 1 {
		//@todo need to review for return type 2024-12-06 15:47:23
		return mo.None[string]()
	}
	matched := mo.TupleToResult(regexp.MatchString(urlPattern, plugin.Url))
	if matched.IsError() || !matched.MustGet() {
		//@todo need to review for return type 2024-12-06 15:47:27
		return mo.None[string]()
	}
	return mo.TupleToOption(lo.Last(strings.Split(plugin.Url, "@")))
}

func (plugin Plugin) validate() mo.Result[string] {
	ver := plugin.version()
	if ver.IsAbsent() && len(plugin.Url) == 0 {
		return mo.Ok[string]("")
	}
	suffix := strings.ReplaceAll(ver.MustGet(), ".", "_")
	return lo.IfF(env.WindowsEnv(), func() mo.Result[string] {
		return mo.Ok(fmt.Sprintf("%s_%s.exe", plugin.Name, suffix))
	}).Else(mo.Ok(fmt.Sprintf("%s_%s", plugin.Name, suffix)))
}

func (plugin Plugin) setup() error {
	if binary := plugin.validate(); binary.IsError() {
		return binary.Error()
	}
	// update build.yaml
	pName := fmt.Sprintf("%s.%s", pluginsKey, plugin.Name)
	argKey := fmt.Sprintf("%s.%s", pName, "args")
	urlKey := fmt.Sprintf("%s.%s", pName, "url")
	if builder().Get(argKey) == nil {
		values := map[string]any{
			argKey: plugin.Args,
			urlKey: plugin.Url,
			fmt.Sprintf("%s.%s", pName, "description"): plugin.Description,
		}
		if err := updateCfg(values); err != nil {
			return err
		}
	}
	// set plugin specific configuration
	if len(plugin.Config) > 0 {
		name := filepath.Join(RootDir(), plugin.Config)
		if rs := mo.TupleToResult(os.Stat(name)); rs.IsError() {
			if source := mo.TupleToResult(resources.Open(plugin.Config)); source.IsOk() {
				defer source.MustGet().Close()
				if dest := mo.TupleToResult(os.Create(name)); dest.IsOk() {
					defer dest.MustGet().Close()
					io.Copy(dest.MustGet(), source.MustGet())
				}
			}
		}
	}
	// setup shell
	if len(plugin.Shell) > 0 {
		name := strings.FieldsFunc(plugin.Name, func(r rune) bool { return r == '.' })[1]
		dest := mo.TupleToResult(os.Create(filepath.Join(hookDir(plugin.Name), name)))
		if dest.IsError() {
			return dest.Error()
		}
		file := dest.MustGet()
		defer file.Close()
		buf := bufio.NewWriter(file)
		buf.WriteString(lo.If(env.WindowsEnv(), "#!/usr/bin/env bash\n").Else("#!/bin/sh\n"))
		buf.WriteString(fmt.Sprintf("\n%s", plugin.Shell))
		buf.Flush()
		return nil
	}
	return nil
}
func hookDir(hookName string) string {
	dir := lo.If(strings.HasPrefix(hookName, "git"), filepath.Join(".git", "hooks")).
		Else("")
	return lo.IfF(env.Active().Test(), func() string {
		mock := filepath.Join(TargetDir(), dir)
		if _, err := os.Stat(mock); err != nil {
			os.MkdirAll(mock, os.ModePerm) //nolint
		}
		return mock
	}).Else(filepath.Join(RootDir(), dir))
}

// download plugin
func (plugin Plugin) download() error {
	binary := plugin.validate()
	if binary.IsError() {
		return binary.Error()
	} else if len(binary.MustGet()) == 0 {
		return nil
	}
	gopath := GoPath()
	if rs := mo.TupleToResult(os.Stat(filepath.Join(gopath, binary.MustGet()))); rs.IsOk() {
		return nil
	}
	tempGoPath := temporaryGoPath()
	defer os.RemoveAll(tempGoPath)
	cmd := exec.Command("go", "install", plugin.Url) //nolint:gosec
	cmd.Env = lo.Map(os.Environ(), func(pair string, _ int) string {
		if strings.HasPrefix(pair, "GOPATH=") {
			return fmt.Sprintf("%s=%s", "GOPATH", tempGoPath)
		}
		return pair
	})
	if err := PtyCmdOutput(cmd, fmt.Sprintf("install %s", plugin.Url), TargetDir(), nil); err != nil {
		return err
	}
	return filepath.WalkDir(tempGoPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(filepath.Dir(path), "bin") && strings.HasPrefix(d.Name(), plugin.Name) {
			err = os.Rename(path, filepath.Join(gopath, binary.MustGet()))
			if err != nil {
				return err
			}
		} else {
			// change the permission for deletion
			os.Chmod(path, 0o766) //nolint
		}
		return err
	})
}

func (plugin Plugin) Install() error {
	if err := plugin.setup(); err != nil {
		return err
	}
	return plugin.download()
}

func (plugin Plugin) Execute() error {
	binary := plugin.validate()
	if binary.IsError() {
		return binary.Error()
	}
	// hook
	if len(binary.MustGet()) == 0 {
		for _, arg := range plugin.Args {
			op := PluginByN(arg)
			if op.IsAbsent() {
				return fmt.Errorf("can not find plugin %s", arg)
			}
			return op.MustGet().Execute()
		}
	}
	abs := filepath.Join(GoPath(), binary.MustGet())
	if rs := mo.TupleToResult(os.Stat(abs)); rs.IsError() {
		if err := plugin.download(); err != nil {
			return err
		}
	}
	// always use absolute path
	pCmd := exec.Command(abs, plugin.Args...) //nolint #gosec
	if err := PtyCmdOutput(pCmd, fmt.Sprintf("start %s", plugin.Name), TargetDir(), nil); err != nil {
		return err
	}
	if pCmd.ProcessState.ExitCode() != 0 {
		return fmt.Errorf("faild %d", pCmd.ProcessState.ExitCode())
	}
	return nil
}

func PluginByN(name string) mo.Option[Plugin] {
	return mo.TupleToOption(lo.Find(Plugins(), func(p Plugin) bool {
		return lo.If(len(p.Url) == 0, p.Name == name).Else(strings.HasSuffix(p.Name, name))
	}))
}
