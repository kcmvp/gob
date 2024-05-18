package artifact

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/fatih/color"
	"github.com/samber/lo"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

const modulePattern = `^[^@]+@?[^@\s]+$`

type Plugin struct {
	Alias       string `json:"alias" mapstructure:"alias"`
	Args        string `json:"args" mapstructure:"args"`
	Url         string `json:"url" mapstructure:"url"` //nolint
	Config      string `json:"load" mapstructure:"load"`
	Description string `json:"description" mapstructure:"description"`
	version     string
	name        string
	module      string
}

func (plugin *Plugin) init() error {
	re := regexp.MustCompile(modulePattern)
	if !re.MatchString(plugin.Url) {
		return fmt.Errorf("invalud tool url %s", plugin.Url)
	}
	plugin.version = "latest"
	plugin.Url = strings.TrimSpace(plugin.Url)
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
		plugin.version = LatestVersion(plugin.module)[0].B
	}
	return nil
}

func LatestVersion(modules ...string) []lo.Tuple2[string, string] {
	modules = lo.Map(modules, func(item string, _ int) string {
		return fmt.Sprintf("%s@latest", item)
	})
	output, _ := exec.Command("go", append([]string{"list", "-m"}, modules...)...).CombinedOutput() //nolint
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	var tuple []lo.Tuple2[string, string]
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		entry := strings.Split(line, " ")
		tuple = append(tuple, lo.Tuple2[string, string]{A: entry[0], B: entry[1]})
	}
	return tuple
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

func NewPlugin(url string) (Plugin, error) {
	plugin := Plugin{
		Url: url,
	}
	if err := plugin.init(); err != nil {
		return Plugin{}, err
	}
	return plugin, nil
}
func (plugin Plugin) Version() string {
	return plugin.version
}

func (plugin Plugin) Name() string {
	return plugin.name
}

func (plugin Plugin) taskName() string {
	return lo.If(len(plugin.Alias) > 0, plugin.Alias).Else(plugin.Name())
}

func (plugin Plugin) Binary() string {
	return lo.IfF(Windows(), func() string {
		return fmt.Sprintf("%s-%s.exe", plugin.Name(), plugin.Version())
	}).Else(fmt.Sprintf("%s-%s", plugin.Name(), plugin.Version()))
}

// install a plugin when it does not exist
func (plugin Plugin) install() (string, error) {
	gopath := GoPath()
	if _, err := os.Stat(filepath.Join(gopath, plugin.Binary())); err == nil {
		return "", nil
	}
	tempGoPath := temporaryGoPath()
	defer os.RemoveAll(tempGoPath)
	fmt.Printf("Installing %s@%s to %s ...... \n", plugin.Url, plugin.Version(), gopath)
	cmd := exec.Command("go", "install", fmt.Sprintf("%s@%s", plugin.Url, plugin.Version())) //nolint:gosec
	cmd.Env = lo.Map(os.Environ(), func(pair string, _ int) string {
		if strings.HasPrefix(pair, "GOPATH=") {
			return fmt.Sprintf("%s=%s", "GOPATH", tempGoPath)
		}
		return pair
	})
	if data, err := cmd.CombinedOutput(); err != nil {
		return tempGoPath, errors.New(color.RedString(string(data)))
	}
	return tempGoPath, filepath.WalkDir(tempGoPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(filepath.Dir(path), "bin") && strings.HasPrefix(d.Name(), plugin.name) {
			err = os.Rename(path, filepath.Join(gopath, plugin.Binary()))
			if err != nil {
				return err
			}
			fmt.Printf("%s is installed successfully \n", plugin.Url)
		} else {
			// change the permission for deletion
			os.Chmod(path, 0o766) //nolint
		}
		return err
	})
}

func (plugin Plugin) Execute() error {
	if _, err := plugin.install(); err != nil {
		return err
	}
	// always use absolute path
	pCmd := exec.Command(filepath.Join(GoPath(), plugin.Binary()), strings.Split(plugin.Args, " ")...) //nolint #gosec
	if err := StreamCmdOutput(pCmd, plugin.taskName()); err != nil {
		return err
	}
	if pCmd.ProcessState.ExitCode() != 0 {
		return fmt.Errorf("faild %d", pCmd.ProcessState.ExitCode())
	}
	return nil
}
