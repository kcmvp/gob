package builder

import (
	"bytes"
	"errors"
	"log"
	"os"
	"path/filepath"
	"regexp"

	"github.com/fatih/color"
)

func (gitHook *GitHook) commitMessageAfterScan(project *Project, refs ...string) {
	var prettyJSON bytes.Buffer
	err := os.WriteFile(filepath.Join(project.scriptsDir, quality), prettyJSON.Bytes(), os.ModePerm)
	FatalIfError(err)
}

func (gitHook *GitHook) commitMessageBeforeScan(args ...string) error {
	reg, err := regexp.Compile(gitHook.cfg.MsgPattern)
	if err == nil && !reg.MatchString(args[0]) {
		msg := color.RedString("commit message must follow %s", gitHook.cfg.MsgPattern)
		log.Println(color.RedString("commit message must follow %s", gitHook.cfg.MsgPattern))
		err = errors.New(msg)
	}
	return err
}
