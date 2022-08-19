/*
Copyright © 2022 kcmvp <kcheng.mvp@gmail.com>
*/
package cmd

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/kcmvp/gbt/builder"
	"github.com/kcmvp/gbt/builder/githook"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
)

//go:embed template/*.tmpl
var templateDir embed.FS

// githookCmd represents the githook command.
var githookCmd = &cobra.Command{
	Use:   "githook",
	Short: "Generate git hook for currentProject",
	Long:  `Generate git hooks for currentProject, includes: commit_message, pre_push`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		_, err := os.Stat(git.GitDirName)
		if errors.Is(err, os.ErrNotExist) {
			err = fmt.Errorf("currentProject currentProject is not versioned in git: %w", err)
		}
		return err
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return generateHook(cmd.Context())
	},
}

type Hook struct {
	Target string
	Type   string
}

func generateHook(ctx context.Context) error {
	project, _ := ctx.Value(_ctxBuilder).(*builder.Project)
	scriptDir := project.ScriptDir()
	gitDir := project.GirDir()

	for s, g := range githook.Hooks() {
		gof := fmt.Sprintf("%s.go", g)
		abs, _ := filepath.Abs(filepath.Join(scriptDir, gof))
		tf, err := templateDir.ReadFile(filepath.Join("template", fmt.Sprintf("%s.tmpl", g)))
		if err != nil {
			return err
		}
		generateFile(string(tf), abs, nil, false)
		tf, err = templateDir.ReadFile(filepath.Join("template", "hook.tmpl"))
		if err != nil {
			return err
		}
		generateFile(string(tf), filepath.Join(gitDir, "hooks", s), Hook{abs, s}, true)
	}
	return nil
}

func init() {
	setupCmd.AddCommand(githookCmd)
}
