package builder

import (
	"encoding/json"
	"fmt"
	"github.com/fatih/color"
	"github.com/go-git/go-git/v5"
	"log"
	"os"
	"path/filepath"
)

func (gitHook *GitHook) prePushBeforeScan() error {
	w, err := gitHook.rep.Worktree()
	//common.FatalIfError(err)
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
	FatalIfError(err)
	previous := Quality{}
	err = json.Unmarshal(data, &previous)
	FatalIfError(err)

	cm := project.quality.Coverage.Method
	cl := project.quality.Coverage.Line
	pm := previous.Coverage.Method
	pl := previous.Coverage.Line

	if cm != pm || cl != pl {
		log.Fatalln(color.RedString("test coverage decrease method : %f -> %f, line: %f -> %s", pm, cm, pl, cl))
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
