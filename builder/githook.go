package builder

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"

	"github.com/fatih/color"
	"github.com/go-git/go-git/v5"
	"github.com/kcmvp/gob/boot"
)

type hookData struct {
	Target string
	Type   string
}

type GitErr struct {
	Msg string
}

func (g GitErr) Error() string {
	return g.Msg
}

func genGitHooks(gitHome, scriptDir string) error {
	var err error
	var tf []byte
	if _, err = git.PlainOpen(gitHome); err != nil {
		return &GitErr{fmt.Sprintf("project is not at version control:%s", err.Error())}
	}
	for k, v := range HookMap() {
		g := fmt.Sprintf("%s.go", k)
		abs, _ := filepath.Abs(filepath.Join(scriptDir, g))
		// @todo code refactor for unit test
		if _, err = os.Stat(abs); errors.Is(err, os.ErrNotExist) {
			tf, err = templateDir.ReadFile(filepath.Join("template", fmt.Sprintf("%s.tmpl", k)))
			if err != nil {
				return err
			}
			err = boot.GenerateFile(string(tf), abs, nil, false)
			if err != nil {
				return err //nolint
			}
		}
		tf, err = templateDir.ReadFile(filepath.Join("template", "hook.tmpl"))
		if err != nil {
			return err //nolint
		}
		err = boot.GenerateFile(string(tf), filepath.Join(gitHome, "hooks", v), hookData{abs, v}, true)
		if err != nil {
			return err
		}
	}
	return err
}

func validateCommitMsg(msg, pattern string) error {
	rep := regexp.MustCompile(`\r?\n`)
	commitMsg := rep.ReplaceAllString(msg, "")
	reg, err := regexp.Compile(pattern)
	if err != nil {
		return err //nolint:wrapcheck
	} else if !reg.MatchString(commitMsg) {
		return fmt.Errorf("commit message must follow %s", pattern) //nolint
	}
	return err //nolint
}

/*
func GitCheckout(files ...string) {
	w, _ := gitHooker.repo.Worktree()
	s, _ := w.Status()
	for _, file := range files {
		log.Printf("git status: %s: %s\n", file, string(s.File(file).Worktree))

		if err := w.AddWithOptions(&git.AddOptions{Path: file}); err != nil {
			log.Println(color.RedString("git add error:%s", err.Error()))
		}
	}
}

func GitAdd(files ...string) {
	w, _ := gitHooker.repo.Worktree()
	s, _ := w.Status()
	for _, file := range files {
		log.Printf("git status: %s: %s\n", file, string(s.File(file).Worktree))

		if err := w.AddWithOptions(&git.AddOptions{Path: file}); err != nil {
			log.Println(color.RedString("git add error:%s", err.Error()))
		}
	}
}
*/

func PrePush(version, target string, repo *git.Repository) {
	input, _ := os.ReadFile(os.Args[1])
	rep := regexp.MustCompile(`\r?\n`)
	commitMsg := rep.ReplaceAllString(string(input), "")
	reg, err := regexp.Compile(version)
	if err == nil && !reg.MatchString(commitMsg) {
		log.Fatalln(color.RedString("commit message must follow %s", version))
	}
	// check the consistent between version and target
}
