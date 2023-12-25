package shared

import (
	"context"
	"errors"
	"fmt"
	"github.com/samber/lo"
	"os/exec"
	"strings"
	"time"
)

// LatestVersion  return the latest version of the tool by module name and version filter(eg 'v1.5.*').
// it will return the latest version if success otherwise return an error
// This function may fail due the fact the domain 'https://github.com' is not accessible,
func LatestVersion(module, filter string) (string, error) {
	ch := make(chan string, 1)
	defer close(ch)
	parts := strings.Split(module, "/")
	module = strings.Join(parts[0:3], "/")
	url := fmt.Sprintf("https://%s.git", module)
	ctx, _ := context.WithDeadline(context.Background(), time.Now().Add(10*time.Second))
	args := []string{"ls-remote", "--sort=-version:refname", "--tags"}
	args = append(args, url)
	if len(filter) > 0 {
		args = append(args, filter)
	}
	cmd := exec.CommandContext(ctx, "git", args...)
	go func() {
		data, err := cmd.CombinedOutput()
		if err == nil {
			v, _ := lo.Last(strings.Split(strings.Split(string(data), "\n")[0], "/"))
			ch <- v
		}
	}()
	var ver string
	select {
	case <-ctx.Done():
		cmd.Process.Kill()
	case ver = <-ch:
	}
	return ver, lo.IfF(ctx.Err() != nil, func() error {
		return ctx.Err()
	}).ElseIfF(len(ver) == 0, func() error {
		return errors.New("failed to get module version")
	}).Else(nil)
}
