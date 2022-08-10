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

func (gitHook *GitHook) commitMessageAfterScan(project *Project) error {
	var prettyJSON bytes.Buffer
	return os.WriteFile(filepath.Join(project.scriptsDir, quality), prettyJSON.Bytes(), os.ModePerm)
}

func (gitHook *GitHook) commitMessageBeforeScan(args ...string) error {
	reg, err := regexp.Compile(gitHook.cfg.MsgPattern)
	rep := regexp.MustCompile(`\r?\n`)
	msg := rep.ReplaceAllString(args[0], "")
	if err == nil && !reg.MatchString(msg) {
		msg := color.RedString("commit message must follow %s", gitHook.cfg.MsgPattern)
		log.Println(color.RedString("commit message must follow %s", gitHook.cfg.MsgPattern))
		err = errors.New(msg)
	}
	return err
}
