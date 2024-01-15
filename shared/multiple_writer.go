package shared

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/samber/lo"

	"github.com/creack/pty"
)

func StreamCmdOutput(cmd *exec.Cmd, file string) error {
	// Start the command with a pty
	var scanner *bufio.Scanner
	if ptmx, err := pty.Start(cmd); err == nil {
		scanner = bufio.NewScanner(ptmx)
		defer ptmx.Close()
	} else {
		color.Yellow("device not support tty will fall back plain model")
		if rd, err := cmd.StdoutPipe(); err == nil {
			scanner = bufio.NewScanner(rd)
		} else {
			return err
		}
	}

	// Create a file to save the output
	log, err := os.Create(file)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return err
	}
	defer log.Close()

	// Create a regular expression to match color escape sequences
	colorRegex := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	// Goroutine to remove color escape sequences, print the colored output, and write the modified output to the file
	ch := make(chan string)
	eof := false
	go func() {
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if len(line) > 0 {
				ch <- line
			}
		}
		eof = true
		if err = scanner.Err(); err != nil {
			fmt.Println("Error reading output:", err)
		}
	}()
	ticker := time.NewTicker(150 * time.Millisecond)
	inProgress := false
	progress := NewProgress()
	for !eof {
		select {
		case line := <-ch:
			progress.Reset()
			lineWithoutColor := colorRegex.ReplaceAllString(line, "")
			_, err = log.WriteString(lineWithoutColor + "\n")
			line = lo.IfF(inProgress, func() string {
				inProgress = false
				return fmt.Sprintf("\r%-20s", line)
			}).Else(line)
			fmt.Println(line)
			if err != nil {
				fmt.Println("Error writing to file:", err)
				break
			}
		case <-ticker.C:
			if !inProgress {
				inProgress = true
			}
			_ = progress.Add(1)
		}
	}
	_ = progress.Finish()
	ticker.Stop()
	return cmd.Wait()
}
