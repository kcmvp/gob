package internal

import (
	"bufio"
	"fmt"
	"github.com/fatih/color"
	"github.com/go-git/go-git/v5"
	"github.com/samber/lo"
	"os"
	"path/filepath"
	"strings"
)

const (
	GitHookKey = "githook"
	CommitMsg  = "commit-msg"
	PreCommit  = "pre-commit"
	PrePush    = "pre-push"
)

var hookMap = map[string]string{
	CommitMsg: "gb exec commitMsg $1 $2",
	PreCommit: "gb exec preCommit",
	PrePush:   "gb exec prePush",
}

type GitHook struct {
	CommitMsg string   `mapstructure:"commit-msg"`
	PreCommit []string `mapstructure:"pre-commit"`
	PrePush   []string `mapstructure:"pre-push"`
}

func (project *Project) GitHook() GitHook {
	var hook GitHook
	project.viper.UnmarshalKey(GitHookKey, &hook)
	return hook
}

func (project *Project) SetupHook(create bool) {
	// generate configuration
	gitHook := CurProject().GitHook()
	noHook := len(strings.TrimSpace(gitHook.CommitMsg)) == 0 && len(gitHook.PreCommit) == 0 &&
		len(gitHook.PrePush) == 0
	if create && noHook {
		hook := map[string]any{
			fmt.Sprintf("%s.%s", GitHookKey, CommitMsg): "^#[0-9]+:\\s*.{10,}$",
			fmt.Sprintf("%s.%s", GitHookKey, PreCommit): []string{"lint", "test"},
			fmt.Sprintf("%s.%s", GitHookKey, PrePush):   []string{"lint", "test"},
		}
		project.viper.MergeConfigMap(hook)
		project.viper.WriteConfigAs(project.Configuration())
	}
	// always generate hook script
	if _, err := git.PlainOpen(CurProject().Root()); err != nil {
		if create {
			color.Yellow("Project is not in the source control")
		}
		return
	}
	hooks := lo.If(create, lo.MapToSlice(hookMap, func(key string, _ string) string {
		return key
	})).ElseF(func() []string {
		var scripts []string
		if len(gitHook.CommitMsg) > 0 {
			scripts = append(scripts, CommitMsg)
		}
		if len(gitHook.PreCommit) > 0 {
			scripts = append(scripts, PreCommit)
		}
		if len(gitHook.PrePush) > 0 {
			scripts = append(scripts, PrePush)
		}
		return scripts
	})
	shell := lo.IfF(Windows(), func() string {
		return "#!/usr/bin/env pwsh\n"
	}).Else("#!/bin/sh\n")
	hookDir := filepath.Join(CurProject().Root(), ".git", "hooks")
	for name, script := range hookMap {
		if lo.Contains(hooks, name) {
			msgHook, _ := os.OpenFile(filepath.Join(hookDir, name), os.O_RDWR|os.O_CREATE|os.O_TRUNC, os.ModePerm)
			writer := bufio.NewWriter(msgHook)
			writer.WriteString(shell)
			writer.WriteString("\n")
			writer.WriteString(script)
			writer.Flush()
			msgHook.Close()
		} else {
			os.Remove(filepath.Join(hookDir, name))
		}
	}
}
