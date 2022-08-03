/*
Copyright © 2022 kcmvp <kcheng.mvp@gmail.com>
*/
package cmd

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

//go:embed template
var templateDir embed.FS

var _projectRoot = "projectRoot"

var hookMap = map[string]string{"commit-msg": "message_hook.go",
	"pre-push": "push_hook.go"}

// githookCmd represents the githook command
var githookCmd = &cobra.Command{
	Use:   "githook",
	Short: "Generate git hook for project",
	Long:  `Generate git hooks for project, includes: commit_message, pre_push`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		dir, _ := os.Getwd()
		var err error
		for dir != string(os.PathSeparator) {
			if _, err = os.Stat(filepath.Join(dir, ".git")); err != nil {
				dir = filepath.Dir(dir)
			} else {
				ctx := context.WithValue(cmd.Context(), _projectRoot, dir)
				cmd.SetContext(ctx)
			}
		}
		if err != nil {
			err = fmt.Errorf("current project is not versioned in git")
		}
		return err
	},
	Run: func(cmd *cobra.Command, args []string) {
		generateHook(cmd.Context())
	},
}

func generateHook(ctx context.Context) {
	root := ctx.Value(_projectRoot).(string)
	dir := filepath.Join(root, scriptDir)
	os.MkdirAll(dir, os.ModePerm)
	for k, v := range hookMap {
		hook := filepath.Join(root, ".git", "hooks", k)
		if f, err := os.OpenFile(hook, os.O_RDWR|os.O_CREATE|os.O_EXCL, os.ModePerm); err == nil {
			fmt.Println(fmt.Sprintf("generate %s hook", k))
			f.WriteString("#!/bin/sh\n\n")
			f.WriteString(fmt.Sprintf("go run %s $1 $2 -event=%s\n", filepath.Join(dir, v), k))
			f.Close()
		} else if errors.Is(err, os.ErrExist) {
			fmt.Println(fmt.Sprintf("%s exists", hook))
		}
		tn := strings.Replace(v, ".go", ".tmpl", 1)
		if data, err := templateDir.ReadFile(tn); err == nil {
			generateFile(string(data), filepath.Join(dir, v), nil)
		}
	}

}

func init() {
	rootCmd.AddCommand(githookCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// githookCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// githookCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
