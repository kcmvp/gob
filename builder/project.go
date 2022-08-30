package builder

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
)

type Buildable interface {
	ScriptDir() string
	TargetDir() string
	RootDir() string
	GitHome() (string, error)
}

var _ Buildable = (*DefaultBuildable)(nil)

type DefaultBuildable struct {
	scriptDir, targetDir, root string
}

func (d *DefaultBuildable) GitHome() (string, error) {
	dir := filepath.Join(d.RootDir(), git.GitDirName)
	_, err := os.Stat(filepath.Join(d.RootDir(), git.GitDirName))
	if err != nil {
		dir = ""
		err = fmt.Errorf("invalid git repository:%w", err)
	}
	return dir, err
}

func (d *DefaultBuildable) ScriptDir() string {
	return d.scriptDir
}

func (d *DefaultBuildable) TargetDir() string {
	return d.targetDir
}

func (d *DefaultBuildable) RootDir() string {
	return d.root
}

func NewDefaultBuildable(root string) *DefaultBuildable {
	return &DefaultBuildable{
		root:      root,
		scriptDir: filepath.Join(root, projectScriptDir),
		targetDir: filepath.Join(root, projectTargetDir),
	}
}
