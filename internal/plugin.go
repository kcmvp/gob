package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/kcmvp/gob/cmd/shared" //nolint
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/fatih/color"
	"github.com/samber/lo"
)

const modulePattern = `^[^@]+@?[^@\s]+$`

type Plugin struct {
	Alias   string `json:"alias" mapstructure:"alias"`
	Command string `json:"command" mapstructure:"command"`
	Args    string `json:"args" mapstructure:"args"`
	Url     string `json:"url" mapstructure:"url"` //nolint
	Config  string `json:"config" mapstructure:"config"`
	version string
	name    string
	module  string
}

func (plugin *Plugin) init() error {
	re := regexp.MustCompile(modulePattern)
	if !re.MatchString(plugin.Url) {
		return fmt.Errorf("invalud tool url %s", plugin.Url)
	}
	plugin.version = "latest"
	reg := regexp.MustCompile(`@\S*`)
	matches := reg.FindAllString(plugin.Url, -1)
	if len(matches) > 0 {
		plugin.version = strings.Trim(matches[0], "@")
	}
	plugin.Url = reg.ReplaceAllString(plugin.Url, "")
	plugin.module = plugin.Url
	if strings.Contains(plugin.module, "github.com") {
		segs := strings.Split(plugin.module, "/")
		if len(segs) > 2 {
			plugin.module = strings.Join(segs[0:3], "/")
		}
	}
	plugin.name, _ = lo.Last(strings.Split(plugin.module, "/"))
	if plugin.version == "latest" {
		output, err := exec.Command("go", "list", "-m", fmt.Sprintf("%s@latest", plugin.module)).CombinedOutput() //nolint:gosec
		if err != nil {
			return fmt.Errorf("failed to get version of %s", plugin.Url)
		}
		plugin.version = strings.Fields(strings.ReplaceAll(string(output), "\n", ""))[1]
	}
	return nil
}

func (plugin Plugin) Module() string {
	return plugin.module
}

func (plugin *Plugin) UnmarshalJSON(data []byte) error {
	type Embedded Plugin
	aux := &struct {
		*Embedded
	}{
		(*Embedded)(plugin),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		fmt.Println(err.Error())
		return err
	}
	return (*Plugin)(aux.Embedded).init()
}

func NewPlugin(url string, options ...string) (Plugin, error) {
	plugin := Plugin{Url: url}
	for i, option := range options {
		if i == 2 {
			plugin.Command = option
		} else if i == 3 {
			plugin.Args = option
		}
	}
	if err := plugin.init(); err != nil {
		return Plugin{}, err
	}
	return plugin, nil
}

func (plugin Plugin) Equal(other Plugin) bool {
	return strings.Split(plugin.Url, "@")[0] == strings.Split(other.Url, "@")[0]
}

func (plugin Plugin) Version() string {
	return plugin.version
}

func (plugin Plugin) Name() string {
	return plugin.name
}

func (plugin Plugin) Binary() string {
	return lo.IfF(Windows(), func() string {
		return fmt.Sprintf("%s-%s.exe", plugin.Name(), plugin.Version())
	}).Else(fmt.Sprintf("%s-%s", plugin.Name(), plugin.Version()))
}

func (plugin Plugin) setup() error {
	for k, v := range map[string]string{
		"alias":   plugin.Alias,
		"command": plugin.Command,
		"args":    plugin.Args,
		"url":     fmt.Sprintf("%s@%s", plugin.Url, plugin.Version()),
	} {
		if len(v) > 0 {
			project.viper.Set(fmt.Sprintf("%s.%s.%s", pluginKey, plugin.Name(), k), v)
		}
	}
	if err := project.viper.WriteConfigAs(project.Configuration()); err != nil {
		return errors.New(color.RedString(err.Error()))
	}
	return nil
}

// install a plugin when it does not exist
func (plugin Plugin) install() error {
	gopath := GoPath()
	if _, err := os.Stat(filepath.Join(gopath, plugin.Binary())); err == nil {
		return nil
	}
	tempGoPath := TemporaryGoPath()
	fmt.Printf("Installing %s ...... \n", fmt.Sprintf("%s@%s", plugin.Url, plugin.Version()))
	cmd := exec.Command("go", "install", fmt.Sprintf("%s@%s", plugin.Url, plugin.Version())) //nolint:gosec
	cmd.Env = lo.Map(os.Environ(), func(pair string, _ int) string {
		if strings.HasPrefix(pair, "GOPATH=") {
			return fmt.Sprintf("%s=%s", "GOPATH", tempGoPath)
		}
		return pair
	})
	if err := cmd.Run(); err != nil {
		return errors.New(color.RedString("failed to install the plugin %s:%s", plugin.Url, err.Error()))
	}
	defer func() {
		os.RemoveAll(tempGoPath)
	}()

	return filepath.WalkDir(tempGoPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasPrefix(d.Name(), plugin.name) {
			err = os.Rename(path, filepath.Join(gopath, plugin.Binary()))
			if err != nil {
				return err
			}
			fmt.Printf("%s is installed successfully \n", plugin.Url)
			return filepath.SkipAll
		}
		return err
	})
}

func (plugin Plugin) Execute() error {
	exeCmd := exec.Command(plugin.Binary(), strings.Split(plugin.Args, " ")...) //nolint #gosec
	return shared.StreamExtCmdOutput(exeCmd, filepath.Join(CurProject().Target(), fmt.Sprintf("%s.log", plugin.Name())))
}
