// Package cmd /*
package command

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"github.com/fatih/color"
	"github.com/kcmvp/gob/cmd/gbc/artifact"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"os" //nolint
	"path/filepath"
	"strings" //nolint
	"sync"    //nolint
)

const resourceDir = "resources"

//go:embed resources/*
var resources embed.FS

var (
	once     sync.Once
	template string
)

func usageTemplate() string {
	once.Do(func() {
		bytes, _ := resources.ReadFile(filepath.Join(resourceDir, "usage.tmpl"))
		template = color.YellowString(string(bytes))
	})
	return template
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
		return artifact.CurProject().Validate()
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
	rootCmd.Example = rootExample()
	rootCmd.SetUsageTemplate(usageTemplate())
	rootCmd.SetErrPrefix(color.RedString("Error:"))
	rootCmd.SetFlagErrorFunc(func(command *cobra.Command, err error) error {
		return lo.IfF(err != nil, func() error {
			return fmt.Errorf(color.RedString(err.Error()))
		}).Else(nil)
	})
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
