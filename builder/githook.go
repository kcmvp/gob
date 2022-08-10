package builder

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/go-git/go-git/v5"
)

type HookEvent string

const (
	None          HookEvent = ""
	CommitMessage HookEvent = "commit-msg"
	PrePush       HookEvent = "pre-push"
	//DefaultMsgP             = `#[0-9]{1,7}:\s?\S{10,}`
	// commit message contains two parts which separated by a ':'
	// the first part is a '#' followed by numbers, the second part is the message;
	// in most case you just need to adjust those numbers to satisfy your requirement
	DefaultMsgP = `#[0-9]{1,7}:\s?(\S+\s?){10,}`
)

var HookMap = map[HookEvent]string{
	CommitMessage: "commit_message.go",
	PrePush:       "pre_push.go",
}

type HookCfg struct {
	MsgPattern  string
	MinCoverage float64
	MaxCoverage float64
}

func DefaultHookCfg() *HookCfg {
	return &HookCfg{
		MsgPattern:  DefaultMsgP,
		MinCoverage: 0.40,
		MaxCoverage: 0.85,
	}
}

type GitHook struct {
	rep     *git.Repository
	rootDir string
	cfg     *HookCfg
	event   HookEvent
}

func newGitHook(path string, cfg *HookCfg) *GitHook {
	rep, err := git.PlainOpen(path)
	if err != nil {
		log.Println(color.YellowString("%s is not valid git project", path))
	}
	return &GitHook{
		rep:     rep,
		cfg:     cfg,
		rootDir: path,
		event:   None,
	}
}

func (gitHook *GitHook) validate() {
	if gitHook.rep == nil {
		log.Println(color.YellowString("ignore validate as %s is not valid repository", gitHook.rootDir))
	}
	for h, c := range HookMap {
		hf := filepath.Join(gitHook.rootDir, git.GitDirName, "hooks", string(h))
		c = filepath.Join(gitHook.rootDir, scriptDir, c)
		if _, err := os.Stat(c); err != nil {
			log.Fatalln(color.RedString("can not find %s, run command 'gbt githook' to initialize the hook", c))
		}
		command := fmt.Sprintf(scriptLine, c)
		if lines, err := os.ReadFile(hf); err != nil || !strings.Contains(string(lines), command) {
			if f, err := os.Create(c); err == nil {
				f.WriteString("#!/bin/sh\n\n")
				f.WriteString(fmt.Sprintf("go run %s $1 $2\n", c))
				f.Close()
			} else {
				log.Fatalln(color.RedString("failed to generate %s", h))
			}
		}
	}
}

func (gitHook *GitHook) beforeScan(args ...string) {
	// 1: make sure there are no uncommitted files
	// 2: make sure the existence of scripts/quality.json and not in stage states
	var err error
	switch gitHook.event {
	case PrePush:
		err = gitHook.prePushBeforeScan()
	case CommitMessage:
		err = gitHook.commitMessageBeforeScan(args...)
	case None:
		//
	default:
		//
	}
	if err != nil {
		os.Exit(1)
	}
}

func (gitHook *GitHook) afterScan(project *Project, args ...string) {
	var err error
	switch gitHook.event {
	case PrePush:
		err = gitHook.prePushAfterScan(project, args...)
	case CommitMessage:
		err = gitHook.commitMessageAfterScan(project)
	case None:

	default:
		//
	}
	if err != nil {
		os.Exit(1)
	}
}
