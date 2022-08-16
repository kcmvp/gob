/*
Copyright © 2022 kcmvp <kcheng.mvp@gmail.com>
*/
package cmd

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"github.com/fatih/color"
	"github.com/go-git/go-git/v5"
	"github.com/kcmvp/gbt/builder"
	"github.com/kcmvp/gbt/builder/githook"
	"github.com/spf13/cobra"
	"log"
	"os"
	"path/filepath"
	"text/template"
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

func generateHook(ctx context.Context) error {
	project, _ := ctx.Value(_ctxProject).(*builder.Project)

	scriptDir := project.ScriptDir()
	gitDir := project.GirDir()

	var err error
	var file *os.File
	var data []byte

	for s, g := range githook.Hooks() {
		v := fmt.Sprintf("%s.go", g)
		abs, _ := filepath.Abs(filepath.Join(scriptDir, v))
		file, err = os.OpenFile(abs, os.O_RDWR|os.O_CREATE|os.O_EXCL, os.ModePerm)
		if err != nil {
			if errors.Is(err, os.ErrExist) {
				log.Println(color.YellowString("%s exists", v))
			} else {
				return fmt.Errorf("failed to create hook %+v", err)
			}
		} else {
			if data, err = templateDir.ReadFile(filepath.Join("template", fmt.Sprintf("%s.tmpl", g))); err == nil {
				t, _ := template.New(g).Parse(string(data))
				if err = t.Execute(file, nil); err != nil {
					return fmt.Errorf("failed to create hook %w, %s", err, s)
				}
			}
		}

		file, err = os.Create(filepath.Join(gitDir, "hooks", s))
		if err != nil {
			return fmt.Errorf("failed to create hook %w, %s", err, s)
		} else {
			// generate hook
			data, _ = templateDir.ReadFile(filepath.Join("template", fmt.Sprintf("%s.tmpl", s)))
			t, _ := template.New(s).Parse(string(data))
			if err = t.Execute(file, abs); err != nil {
				return fmt.Errorf("failed to create hook %w, %s", err, s)
			}
		}
	}
	return nil
}

func init() {
	genCmd.AddCommand(githookCmd)
}
