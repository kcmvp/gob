package project

import (
	"bufio"
	"fmt"
	"github.com/kcmvp/gob/core/env"
	"github.com/kcmvp/gob/core/utils"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/creack/pty"
	"github.com/fatih/color"
)

type consoleFormatter func(msg string) string

func PtyCmdOutput(cmd *exec.Cmd, task string, dir string, formatter consoleFormatter) error {
	// Start the command with a pty
	rc, err := func() (io.ReadCloser, error) {
		if internal.WindowsEnv() {
			r, err := cmd.StdoutPipe()
			if err != nil {
				return r, err
			}
			return r, cmd.Start()
		}
		return pty.Start(cmd)
	}()
	if err != nil {
		return err
	}
	defer rc.Close()
	scanner := bufio.NewScanner(rc)
	color.Green(task)
	var log *os.File
	if len(dir) > 0 {
		log, err = os.Create(filepath.Join(dir, fmt.Sprintf("%s.log", strings.ReplaceAll(task, " ", "_"))))
		if err != nil {
			return fmt.Errorf(color.RedString("Error creating file:", err.Error()))
		}
		defer log.Close()
	}
	// Create a regular expression to match color escape sequences
	colorRegex := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	// Goroutine to remove color escape sequences, print the colored output, and write the modified output to the file
	ch := make(chan string)
	eof := false
	go func() {
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if formatter != nil {
				line = strings.TrimSpace(formatter(line))
			}
			if len(line) > 1 {
				ch <- line
			}
		}
		eof = true
		if err = scanner.Err(); err != nil {
			color.Red("Error reading output: %s", err.Error())
		}
	}()
	overwrite := true
	progress := NewProgress()
	ticker := time.NewTicker(150 * time.Millisecond)
	for !eof {
		select {
		case msg := <-ch:
			progress.Reset()
			if log != nil {
				lineWithoutColor := colorRegex.ReplaceAllString(msg, "")
				_, err = log.WriteString(lineWithoutColor + "\n")
				if err != nil {
					color.Red("Error writing to file: %s", err.Error())
					break
				}
			}
			if overwrite {
				overwrite = false
				fmt.Printf("\r%-15s\n", msg)
			} else {
				fmt.Println(msg)
			}
		case <-ticker.C:
			if !overwrite {
				overwrite = true
			}
			_ = progress.Add(1)
		}
	}
	if testEnv := utils.TestEnv(); testEnv.IsPresent() {
		fmt.Printf("\r%-15s\n", "")
	} else {
		progress.Clear() //nolint
	}
	ticker.Stop()
	return cmd.Wait()
}
