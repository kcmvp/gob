package internal

import (
	"bufio"
	"fmt"
	"github.com/fatih/color"
	"github.com/samber/lo"
	lop "github.com/samber/lo/parallel"
	"os"
	"path/filepath"
	"strings"
)

const (
	ExecKey = "exec"
	//hook script name
	CommitMsg = "commit-msg"
	PreCommit = "pre-commit"
	PrePush   = "pre-push"
)

var CommitMsgCmd = fmt.Sprintf("%s-hook", CommitMsg)
var PreCommitCmd = fmt.Sprintf("%s-hook", PreCommit)
var PrePushCmd = fmt.Sprintf("%s-hook", PrePush)

var HookScripts = map[string]string{
	CommitMsg: fmt.Sprintf("gob exec %s $1", CommitMsgCmd),
	PreCommit: fmt.Sprintf("gob exec %s", PreCommitCmd),
	PrePush:   fmt.Sprintf("gob exec %s", PrePushCmd),
}

type GitHook struct {
	CommitMsg string   `mapstructure:"commit-msg-hook"`
	PreCommit []string `mapstructure:"pre-commit-hook"`
	PrePush   []string `mapstructure:"pre-push-hook"`
}

func (project *Project) GitHook() GitHook {
	var hook GitHook
	project.viper.UnmarshalKey(ExecKey, &hook)
	return hook
}

type Execution struct {
	CmdKey  string
	Actions []string
}

func (project *Project) Executions() []Execution {
	values := project.viper.Get(ExecKey).(map[string]any)
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

// Setup git hooks and re-install missing plugins
func (project *Project) Setup(init bool) {
	// generate configuration
	gitHook := CurProject().GitHook()
	noHook := len(strings.TrimSpace(gitHook.CommitMsg)) == 0 && len(gitHook.PreCommit) == 0 &&
		len(gitHook.PrePush) == 0
	if init && noHook {
		hook := map[string]any{
			fmt.Sprintf("%s.%s", ExecKey, CommitMsgCmd): "^#[0-9]+:\\s*.{10,}$",
			fmt.Sprintf("%s.%s", ExecKey, PreCommitCmd): []string{"lint", "test"},
			fmt.Sprintf("%s.%s", ExecKey, PrePushCmd):   []string{"test"},
		}
		project.viper.MergeConfigMap(hook)
		project.viper.WriteConfigAs(project.Configuration())
	}
	// always generate hook script
	if !InGit() {
		if init {
			color.Yellow("Project is not in the source control")
		}
		return
	}
	hooks := lo.If(init, lo.MapToSlice(HookScripts, func(key string, _ string) string {
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
	for name, script := range HookScripts {
		if lo.Contains(hooks, name) {
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
	// install missing plugins
	lop.ForEach(CurProject().Plugins(), func(plugin lo.Tuple4[string, string, string, string], index int) {
		if !CurProject().PluginInstalled(plugin.D) {
			CurProject().InstallPlugin(plugin.D, plugin.B, plugin.C)
		}
	})
}
