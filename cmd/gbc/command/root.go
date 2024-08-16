// Package cmd /*
package command

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/fatih/color"
	"github.com/kcmvp/gob/cmd/gbc/artifact"
	"github.com/kcmvp/gob/utils"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"
	"os" //nolint
	"path"
	"path/filepath"
	"strings" //nolint
	"sync"    //nolint
)

const (
	resourceDir = "resources"
	tmplDir     = "tmpl"
)

var (
	//go:embed resources/*
	resources embed.FS
	//go:embed tmpl/*
	templates embed.FS
	once      sync.Once
	usage     string
)

func usageTemplate() string {
	once.Do(func() {
		bytes, _ := templates.ReadFile(path.Join(tmplDir, "usage.tmpl"))
		usage = color.YellowString(string(bytes))
	})
	return usage
}

func parseArtifacts(cmd *cobra.Command, args []string, section string) (gjson.Result, error) {
	var result gjson.Result
	var data []byte
	var err error
	if test, uqf := utils.TestCaller(); test {
		data, err = os.ReadFile(filepath.Join(artifact.CurProject().Root(), "target", uqf, "config.json"))
	} else {
		data, err = resources.ReadFile(path.Join(resourceDir, "config.json"))
	}
	if err != nil {
		return result, err
	}
	key := strings.ReplaceAll(cmd.CommandPath(), " ", "_")
	result = gjson.GetBytes(data, fmt.Sprintf("%s.%s", key, section))
	if !result.Exists() {
		result = gjson.GetBytes(data, fmt.Sprintf("%s_%s.%s", key, strings.Join(args, "_"), section))
	}
	return result, nil
}

func installPlugins(cmd *cobra.Command, args []string) error {
	result, err := parseArtifacts(cmd, args, "plugins")
	if result.Exists() {
		var data []byte
		var plugins []artifact.Plugin
		err = json.Unmarshal([]byte(result.Raw), &plugins)
		for _, plugin := range plugins {
			if err = artifact.CurProject().InstallPlugin(plugin); err != nil {
				return err
			}
			if len(plugin.Config) > 0 {
				if _, err = os.Stat(filepath.Join(artifact.CurProject().Root(), plugin.Config)); err != nil {
					if data, err = resources.ReadFile(path.Join(resourceDir, plugin.Config)); err == nil {
						if err = os.WriteFile(filepath.Join(artifact.CurProject().Root(), plugin.Config), data, os.ModePerm); err != nil {
							break
						}
					} else {
						break
					}
				}
			}
		}
		if err != nil {
			return errors.New(color.RedString(err.Error()))
		}
	}
	return err
}

func installDeps(cmd *cobra.Command, args []string) error {
	result, err := parseArtifacts(cmd, args, "deps")
	if err != nil {
		return err
	}
	if result.Exists() {
		var cfgDeps []string
		err = json.Unmarshal([]byte(result.Raw), &cfgDeps)
		for _, dep := range cfgDeps {
			if err := artifact.CurProject().InstallDependency(dep); err != nil {
				return err
			}
		}
	}
	return nil
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:       "gbc",
	Short:     color.GreenString(`Go project boot command line`),
	Long:      color.GreenString(`Go project boot command line`),
	ValidArgs: validBuilderArgs(),
	Args: func(cmd *cobra.Command, args []string) error {
		if err := OnlyValidArgs(cmd, args); err != nil {
			return err
		}
		return MinimumNArgs(1)(cmd, args)
	},
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		currentDir, _ := os.Getwd()
		if artifact.CurProject().Root() != currentDir {
			return fmt.Errorf(color.RedString("Please execute the command in the project root dir"))
		}
		//if err := installPlugins(cmd, args); err != nil {
		//	return err
		//}
		//if err := installDeps(cmd, args); err != nil {
		//	return err
		//}
		return artifact.CurProject().SetupHooks(false)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		for _, arg := range lo.Uniq(args) {
			if err := execute(cmd, arg); err != nil {
				return errors.New(color.RedString("%s \n", err.Error()))
			}
		}
		return nil
	},
}

func rootExample() string {
	format := fmt.Sprintf("  %%-%ds %%s", rootCmd.NamePadding())
	builtIn := lo.Map(lo.Filter(buildActions(), func(item Action, _ int) bool {
		return !strings.Contains(item.A, "_")
	}), func(action Action, _ int) string {
		return fmt.Sprintf(format, action.A, action.C)
	})
	lo.ForEach(artifact.CurProject().Plugins(), func(plugin artifact.Plugin, _ int) {
		if !lo.ContainsBy(builtIn, func(item string) bool {
			return strings.HasPrefix(strings.TrimSpace(item), strings.TrimSpace(plugin.Alias))
		}) {
			builtIn = append(builtIn, fmt.Sprintf(format, plugin.Alias, plugin.Description))
		}
	})
	return strings.Join(builtIn, "\n")
}

func Execute() error {
	ctx := context.Background()
	if err := rootCmd.ExecuteContext(ctx); err != nil {
		return fmt.Errorf(color.RedString(err.Error()))
	}
	return nil
}

func init() {
	// rootCmd.Example = rootExample()
	rootCmd.SetUsageTemplate(usageTemplate())
	rootCmd.SetErrPrefix(color.RedString("Error:"))
	rootCmd.SetFlagErrorFunc(func(command *cobra.Command, err error) error {
		return lo.IfF(err != nil, func() error {
			return fmt.Errorf(color.RedString(err.Error()))
		}).Else(nil)
	})
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
