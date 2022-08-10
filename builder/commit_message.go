package builder

import (
	"fmt"
	"regexp"
)

func (gitHook *GitHook) commitMessageBeforeScan(args ...string) error {
	rep := regexp.MustCompile(`\r?\n`)
	commitMsg := rep.ReplaceAllString(args[0], "")
	reg, err := regexp.Compile(gitHook.cfg.MsgPattern)
	if err == nil && !reg.MatchString(commitMsg) {
		err = fmt.Errorf("commit message must follow %s", gitHook.cfg.MsgPattern)
	}
	return err
}
