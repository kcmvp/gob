package shared

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"regexp"

	"github.com/creack/pty"
)

func StreamCmdOutput(cmd *exec.Cmd, file string) error {
	// Start the command with a pty
	ptmx, err := pty.Start(cmd)
	if err != nil {
		fmt.Println("Error starting command:", err)
		return err
	}
	defer ptmx.Close()

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
	go func() {
		scanner := bufio.NewScanner(ptmx)
		for scanner.Scan() {
			line := scanner.Text()

			fmt.Println(line)
			// Remove color escape sequences from the line
			lineWithoutColor := colorRegex.ReplaceAllString(line, "")
			// Write the modified line to the file
			_, err = log.WriteString(lineWithoutColor + "\n")
			if err != nil {
				fmt.Println("Error writing to file:", err)
				break
			}
		}
		if err = scanner.Err(); err != nil {
			fmt.Println("Error reading output:", err)
		}
	}()
	return cmd.Wait()
}
