package builder

import (
	_ "embed"
	"encoding/json"
	"github.com/fatih/color"
	"github.com/kcmvp/gob/cmd/gob/project"
	"github.com/spf13/cobra"
	"log"
)

var (
	//go:embed resources/plugins.json
	data      []byte
	supported []project.Plugin
)

func PluginCmd() *cobra.Command {
	return &cobra.Command{
		Use:       "plugin",
		Short:     color.GreenString(`Setup build plugins for project`),
		Long:      color.GreenString(`Setup build plugins for project`),
		ValidArgs: []string{"init", "add", "list"},
		Args: func(cmd *cobra.Command, args []string) error {
			err := cobra.OnlyValidArgs(cmd, args)
			if err == nil {
				err = cobra.ExactArgs(1)(cmd, args)
			}
			return err
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "init":
				return installBuiltinPlugins()
			case "add":
				return installPlugin()
			case "list":
				return listPlugin()
			default:
				return nil
			}
		},
	}
}

func init() {
	// read configuration
	if err := json.Unmarshal(data, &supported); err != nil {
		log.Fatalf("failed to parse plugins.json %v", err)
	}
}

func installBuiltinPlugins() error {
	for _, plugin := range supported {
		if err := plugin.Install(); err != nil {
			return err
		}
	}
	return nil
}

func listPlugin() error {
	return nil
}

func installPlugin() error {
	return nil
}
