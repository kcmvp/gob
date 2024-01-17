package internal

import (
	"bufio"
	"fmt"
	"github.com/fatih/color"
	"github.com/samber/lo"
	"os"
	"path/filepath"
)

const (
	execCfgKey = "exec"
	//hook script name
	CommitMsg = "commit-msg"
	PreCommit = "pre-commit"
	PrePush   = "pre-push"
)

func HookScripts() map[string]string {
	return map[string]string{
		CommitMsg: fmt.Sprintf("gob exec %s $1", CommitMsg),
		PreCommit: fmt.Sprintf("gob exec %s", PreCommit),
		PrePush:   fmt.Sprintf("gob exec %s $1 $2", PrePush),
	}
}

type GitHook struct {
	CommitMsg string   `mapstructure:"commit-msg"`
	PreCommit []string `mapstructure:"pre-commit"`
	PrePush   []string `mapstructure:"pre-push"`
}

func (project *Project) GitHook() GitHook {
	var hook GitHook
	project.config().UnmarshalKey(execCfgKey, &hook) //nolint
	return hook
}

type Execution struct {
	CmdKey  string
	Actions []string
}

func (project *Project) Executions() []Execution {
	values := project.config().Get(execCfgKey).(map[string]any)
	return lo.MapToSlice(values, func(key string, v any) Execution {
		var actions []string
		if _, ok := v.(string); ok {
			actions = append(actions, v.(string))
		} else {
			actions = lo.Map(v.([]any), func(item any, _ int) string {
				return fmt.Sprintf("%s", item)
			})
		}
		return Execution{CmdKey: key, Actions: actions}
	})
}

// SetupHooks setup git local hooks for project. force means always update gob.yaml
func (project *Project) SetupHooks(force bool) {
	if force {
		hook := map[string]any{
			fmt.Sprintf("%s.%s", execCfgKey, CommitMsg): "^#[0-9]+:\\s*.{10,}$",
			fmt.Sprintf("%s.%s", execCfgKey, PreCommit): []string{"lint", "test"},
			fmt.Sprintf("%s.%s", execCfgKey, PrePush):   []string{"test"},
		}
		if err := project.mergeConfig(hook); err != nil {
			color.Red("failed to setup hook")
		}
	}
	if !InGit() {
		color.Yellow("project is not in the source control")
		return
	}
	_ = project.config().ReadInConfig()
	gitHook := CurProject().GitHook()
	var hooks []string
	if len(gitHook.CommitMsg) > 0 {
		hooks = append(hooks, CommitMsg)
	}
	if len(gitHook.PreCommit) > 0 {
		hooks = append(hooks, PreCommit)
	}
	if len(gitHook.PrePush) > 0 {
		hooks = append(hooks, PrePush)
	}
	shell := lo.IfF(Windows(), func() string {
		return "#!/usr/bin/env pwsh\n"
	}).Else("#!/bin/sh\n")
	for name, script := range HookScripts() {
		if lo.Contains(hooks, name) || force {
			msgHook, _ := os.OpenFile(filepath.Join(CurProject().HookDir(), name), os.O_RDWR|os.O_CREATE|os.O_TRUNC, os.ModePerm)
			writer := bufio.NewWriter(msgHook)
			writer.WriteString(shell)
			writer.WriteString("\n")
			writer.WriteString(script)
			writer.Flush()
			msgHook.Close()
		} else {
			os.Remove(filepath.Join(CurProject().HookDir(), name))
		}
	}
}
