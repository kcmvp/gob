package infra

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/fatih/color"
	"github.com/go-git/go-git/v5"
)

var (
	gitHook *gitHookService
	once    sync.Once
	//go:embed template/*.tmpl
	templateDir embed.FS
)

const HookScriptDir = "scripts"

type gitHookService struct {
	root  string
	repo  *git.Repository
	hooks []string
	valid bool
}

func NewGitHookService(path string) *gitHookService {
	once.Do(func() {
		repo, err := git.PlainOpen(path)
		if err != nil {
			log.Println(color.YellowString("project is not at version control"))
		}
		gitHook = &gitHookService{
			root:  path,
			hooks: []string{"pre-commit", "commit-msg", "pre-push"},
			repo:  repo,
			valid: err == nil,
		}
	})
	return gitHook
}

func Hooks() map[string]string {
	m := map[string]string{}
	for _, h := range gitHook.hooks {
		m[h] = strings.Replace(h, "-", "_", 1)
	}
	return m
}

type hookData struct {
	Target string
	Type   string
}

func SetupHook(genNew bool) error {
	var err error
	var tf []byte
	if !gitHook.valid {
		msg := "current project is not in git"
		log.Println(color.RedString(msg))
		return errors.New(msg) //nolint
	}
	gitDir := filepath.Join(gitHook.root, ".git")
	for s, g := range Hooks() {
		gof := fmt.Sprintf("%s.go", g)
		abs, _ := filepath.Abs(filepath.Join(HookScriptDir, gof))
		if _, err = os.Stat(abs); err != nil {
			if !genNew {
				continue
			}
			if tf, err = templateDir.ReadFile(filepath.Join("template", fmt.Sprintf("%s.tmpl", g))); err == nil {
				err = GenerateFile(string(tf), abs, nil, false)
			}
		} else if tf, err = templateDir.ReadFile(filepath.Join("template", "hook.tmpl")); err == nil {
			err = GenerateFile(string(tf), filepath.Join(gitDir, "hooks", s), hookData{abs, s}, true)
		}
	}
	if err == nil {
		log.Println("git gitHook is setup successfully")
	}
	return err
}

func CommitMsg(pattern string) {
	input, _ := os.ReadFile(os.Args[1])
	rep := regexp.MustCompile(`\r?\n`)
	commitMsg := rep.ReplaceAllString(string(input), "")
	reg, err := regexp.Compile(pattern)
	if err == nil && !reg.MatchString(commitMsg) {
		log.Fatalln(color.RedString("commit message must follow %s", pattern))
	}
}

func GitCheckout(files ...string) {
	w, _ := gitHook.repo.Worktree()
	s, _ := w.Status()
	for _, file := range files {
		log.Printf("git status: %s: %s\n", file, string(s.File(file).Worktree))

		if err := w.AddWithOptions(&git.AddOptions{Path: file}); err != nil {
			log.Println(color.RedString("git add error:%s", err.Error()))
		}
	}
}

func GitAdd(files ...string) {
	w, _ := gitHook.repo.Worktree()
	s, _ := w.Status()
	for _, file := range files {
		log.Printf("git status: %s: %s\n", file, string(s.File(file).Worktree))

		if err := w.AddWithOptions(&git.AddOptions{Path: file}); err != nil {
			log.Println(color.RedString("git add error:%s", err.Error()))
		}
	}
}

func PrePush(version, target string, repo *git.Repository) {
	input, _ := os.ReadFile(os.Args[1])
	rep := regexp.MustCompile(`\r?\n`)
	commitMsg := rep.ReplaceAllString(string(input), "")
	reg, err := regexp.Compile(version)
	if err == nil && !reg.MatchString(commitMsg) {
		log.Fatalln(color.RedString("commit message must follow %s", version))
	}
	// check the consistent between version and target
	s := Report{}
	t := Report{}
	data, err := os.ReadFile(version)
	if err != nil {
		log.Fatalln(color.RedString("can not open file %s", version))
	}
	err = json.Unmarshal(data, &s)
	if err != nil {
		log.Fatalln(color.RedString("incorrect data format %s", version))
	}

	data, err = os.ReadFile(target)
	if err != nil {
		log.Fatalln(color.RedString("can not open file %s", target))
	}
	err = json.Unmarshal(data, &t)
	if err != nil {
		log.Fatalln(color.RedString("incorrect data format %s", target))
	}

	for k, v := range s.Packages {
		if t.Packages[k] != v {
			log.Fatalln(color.RedString("value of %s is not the same between %s and %s", k, s, t))
		}
	}

	if s.Tests != t.Tests {
		log.Fatalln(color.RedString("number of the test is not the same between %s and %s", s, t))
	}
}
