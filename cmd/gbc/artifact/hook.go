package artifact

import (
	"bufio"
	"fmt"
	"github.com/fatih/color"
	"github.com/samber/lo"
	"os"
	"path/filepath"
)

const (
	command    = "gbc"
	execCfgKey = "exec"
	//hook script name
	CommitMsg = "commit-msg"
	PreCommit = "pre-commit"
	PrePush   = "pre-push"
)

func HookScripts() map[string]string {
	return map[string]string{
		CommitMsg: fmt.Sprintf("%s exec %s $1", command, CommitMsg),
		PreCommit: fmt.Sprintf("%s exec %s", command, PreCommit),
		PrePush:   fmt.Sprintf("%s exec %s $1 $2", command, PrePush),
	}
}

type GitHook struct {
	CommitMsg string   `mapstructure:"commit-msg"`
	PreCommit []string `mapstructure:"pre-commit"`
	PrePush   []string `mapstructure:"pre-push"`
}

func (project *Project) GitHook() GitHook {
	var hook GitHook
	project.load().UnmarshalKey(execCfgKey, &hook) //nolint
	return hook
}

type Execution struct {
	CmdKey  string
	Actions []string
}

func (project *Project) Executions() []Execution {
	values := project.load().Get(execCfgKey)
	if values == nil {
		return []Execution{}
	}
	return lo.MapToSlice(values.(map[string]any), func(key string, v any) Execution {
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
func (project *Project) SetupHooks(force bool) error {
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
		return nil
	}
	// force load configuration again for testing
	_ = project.load().ReadInConfig()
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
		return "#!/usr/bin/env bash\n"
	}).Else("#!/bin/sh\n")
	hookDir := CurProject().HookDir()
	for name, script := range HookScripts() {
		if lo.Contains(hooks, name) || force {
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
	return nil
}
