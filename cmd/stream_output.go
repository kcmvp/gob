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
	// Redirect stderr to stdout
	cmd.Stderr = cmd.Stdout
	// Create a file to save the combined output
	outputFile, err := os.Create(file)
	if err != nil {
		return err
	}
	defer outputFile.Close()
	// Create a writer to write the combined output to the file
	writer := io.MultiWriter(os.Stdout, outputFile)
	// Start the command
	err = cmd.Start()
	if err != nil {
		return err
	}
	// Create a scanner to read the combined output line by line
	scanner := bufio.NewScanner(stdout)
	// Start a goroutine to read and write the output
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
