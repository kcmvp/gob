package cmd

import (
	"bufio"
	"fmt"
	"github.com/kcmvp/gob/internal"
	"github.com/samber/lo"
	"io"
	"os"
	"os/exec"
	"strings"
)

func streamOutput(cmd *exec.Cmd, file string, errWords ...string) error {
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
				internal.Red.Fprintln(os.Stdout, line)
				fmt.Fprintln(outputFile, line)
			} else {
				fmt.Fprintln(writer, line)
			}
		}
	}()
	// Wait for the command to finish
	return cmd.Wait()
}
