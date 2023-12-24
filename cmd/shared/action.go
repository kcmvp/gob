package shared

import (
	"bufio"
	"fmt"
	"github.com/fatih/color"
	"github.com/kcmvp/gb/internal"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Execution func(cmd *cobra.Command, args ...string) error

type CmdAction lo.Tuple2[string, Execution]

func PrintCmd(cmd *cobra.Command, msg string) error {
	if ok, file := internal.TestEnv(); ok {
		// Get the call stack
		outputFile, err := os.Create(filepath.Join(internal.CurProject().Target(), file))
		if err != nil {
			return err
		}
		defer outputFile.Close()
		writer := io.MultiWriter(os.Stdout, outputFile)
		fmt.Fprintln(writer, msg)
	} else {
		cmd.Println(msg)
	}
	return nil
}

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

func execPlugin(cmd *cobra.Command, args ...string) error {
	plugin, ok := lo.Find(internal.CurProject().PluginCommands(), func(plugin lo.Tuple3[string, string, string]) bool {
		return plugin.A == args[0]
	})
	if !ok {
		return fmt.Errorf("plugin %s is not conigured", args[0])
	}
	cmds := strings.Split(plugin.C, ",")
	exeCmd := exec.Command(plugin.B, lo.Map(cmds, func(cmd string, _ int) string {
		return strings.TrimSpace(cmd)
	})...)
	return StreamExtCmdOutput(exeCmd, filepath.Join(internal.CurProject().Target(), fmt.Sprintf("%s.log", args[0])), "")
}

func PluginActions() []CmdAction {
	return lo.Map(internal.CurProject().PluginCommands(), func(plugin lo.Tuple3[string, string, string], _ int) CmdAction {
		return CmdAction{
			A: plugin.A,
			B: execPlugin,
		}
	})
}
