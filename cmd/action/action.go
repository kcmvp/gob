package action

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/kcmvp/gob/internal"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

type (
	Execution func(cmd *cobra.Command, args ...string) error
	CmdAction lo.Tuple2[string, Execution]
)

func StreamExtCmdOutput(cmd *exec.Cmd, file string, errWords ...string) error {
	// Create a pipe to capture the command's combined output
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	cmd.Stderr = cmd.Stdout
	outputFile, err := os.Create(file)
	if err != nil {
		return err
	}
	defer outputFile.Close()
	writer := io.MultiWriter(os.Stdout, outputFile)
	err = cmd.Start()
	if err != nil {
		return err
	}
	scanner := bufio.NewScanner(stdout)
	go func() {
		for scanner.Scan() {
			line := scanner.Text()
			if lo.SomeBy(errWords, func(item string) bool {
				return strings.Contains(line, item)
			}) {
				color.Red(line)
				fmt.Fprintln(outputFile, line)
			} else {
				fmt.Fprintln(writer, line)
			}
		}
	}()
	// Wait for the command to finish
	return cmd.Wait()
}

func ValidBuilderArgs() []string {
	builtIn := lo.Map(builtinActions, func(action CmdAction, _ int) string {
		return action.A
	})
	plugin := lo.Map(internal.CurProject().PluginCommands(), func(t3 lo.Tuple3[string, string, string], _ int) string {
		return t3.A
	})
	return append(builtIn, plugin...)
}

func Execute(cmd *cobra.Command, args ...string) error {
	var exeCmd *exec.Cmd
	if plugin, ok := lo.Find(internal.CurProject().PluginCommands(), func(plugin lo.Tuple3[string, string, string]) bool {
		return plugin.A == args[0]
	}); ok {
		exeCmd = exec.Command(plugin.B, lo.Map(strings.Split(plugin.C, ","), func(cmd string, _ int) string {
			return strings.TrimSpace(cmd)
		})...) // #nosec G204
		stdout, err := exeCmd.StdoutPipe()
		if err != nil {
			return err
		}
		outputFile, err := os.Create(filepath.Join(internal.CurProject().Target(), fmt.Sprintf("%s.log", args[0])))
		if err != nil {
			return err
		}
		defer outputFile.Close()
		err = exeCmd.Start()
		if err != nil {
			return err
		}
		scanner := bufio.NewScanner(stdout)
		go func() {
			for scanner.Scan() {
				line := scanner.Text()
				fmt.Fprintln(outputFile, line)
			}
		}()
		// Wait for the command to finish
		err = exeCmd.Wait()
		if err != nil {
			color.Red("%s report is generated at %s \n", args[0], filepath.Join(internal.CurProject().Target(), fmt.Sprintf("%s.log", args[0])))
		} else {
			color.Green("execute %s successfully on project \n", args[0])
		}
		return err
	}
	if action, ok := lo.Find(builtinActions, func(action CmdAction) bool {
		return action.A == args[0]
	}); ok {
		return action.B(cmd, args...)
	}
	return fmt.Errorf("can not find command %s", args[0])
}

// LatestVersion  return the latest version of the tool by module name and version filter(eg 'v1.5.*').
// it will return the latest version if success otherwise return an error
// This function may fail due the fact the domain 'https://github.com' is not accessible,
func LatestVersion(module, filter string) (string, error) {
	ch := make(chan string, 1)
	defer close(ch)
	parts := strings.Split(module, "/")
	module = strings.Join(parts[0:3], "/")
	url := fmt.Sprintf("https://%s.git", module)
	ctx, _ := context.WithDeadline(context.Background(), time.Now().Add(10*time.Second)) //nolint govet
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
		cmd.Process.Kill() //nolint errcheck
	case ver = <-ch:
	}
	return ver, lo.IfF(ctx.Err() != nil, func() error {
		return ctx.Err()
	}).ElseIfF(len(ver) == 0, func() error {
		return errors.New("failed to get module version")
	}).Else(nil)
}
