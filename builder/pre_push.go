package builder

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/go-git/go-git/v5"
)

func (gitHook *GitHook) prePushBeforeScan() error {
	w, err := gitHook.rep.Worktree()
	if err != nil {
		return err
	}
	f := filepath.Join(scriptDir, quality)
	_, err = w.Filesystem.Lstat(f)

	if err != nil {
		return err
		log.Println(color.RedString("make sure %s is under version control", f))
	}
	s, err := w.Status()
	for path, status := range s {
		if status.Worktree != git.Unmodified || status.Staging != git.Unmodified {
			log.Fatalln(color.RedString("%c%c %s", status.Staging, status.Worktree, path))
		}
	}
	return err
}

func (gitHook *GitHook) prePushAfterScan(project *Project, args ...string) error {
	// @todo check coverage & linter
	// 1: should be the same (scripts/quality.json and target/quality.json)
	// 2: should no degrade in test coverage and linter
	data, err := os.ReadFile(filepath.Join(project.scriptsDir, quality))
	if err != nil {
		return err
	}
	previous := Quality{}
	err = json.Unmarshal(data, &previous)
	if err != nil {
		return err
	}

	cm := project.quality.Coverage.Method
	cl := project.quality.Coverage.Line
	pm := previous.Coverage.Method
	pl := previous.Coverage.Line

	if cm != pm || cl != pl {
		msg := "test coverage changed, please run 'go run scripts/builder.go' to update"
		log.Println(color.RedString(msg))
		return errors.New(msg)
	}
	fmt.Println(args)

	// l, err := gitHook.rep.CommitObject(plumbing.NewHash(refs[1]))
	// common.FatalIfError(err)
	// r, err := gitHook.rep.CommitObject(plumbing.NewHash(refs[3]))
	// common.FatalIfError(err)
	// p, err := r.Patch(l)
	// common.FatalIfError(err)
	// for _, patch := range p.FilePatches() {
	//	f, t := patch.Files()
	//	fmt.Printf(f.Path())
	//}
	return err
}
