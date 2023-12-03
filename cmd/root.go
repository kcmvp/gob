// Package cmd /*
package cmd

import (
	"context"
	"github.com/kcmvp/gob/internal"
	"github.com/spf13/cobra"
	"io"
	"os"
	"os/exec"
)

const proCtx = "project"

var validArgsFun = func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	var validArgs []string
	if cmd.Name() == "gob" {
		validArgs = append(validArgs, []string{"build", "clean", "test", "package"}...)
	}
	return validArgs, cobra.ShellCompDirectiveNoSpace
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:               "gob",
	Short:             "Go project boot",
	Long:              `Supply most frequently usage and best practice for go project development`,
	ValidArgsFunction: validArgsFun,
	//Args: func(cmd *cobra.Cmd, args []string) error {
	//	if err := cobra.MinimumNArgs(1)(cmd, args); err != nil {
	//		return fmt.Errorf(internal.Red.Sprintf(err.Error()))
	//	}
	//	if err := cobra.OnlyValidArgs(cmd, args); err != nil {
	//		return fmt.Errorf(internal.Red.Sprintf(err.Error()))
	//	}
	//	return nil
	//},
	//RunE: func(cmd *cobra.Cmd, args []string) error {
	//	fmt.Println(args)
	//	return nil
	//},
	// Uncomment the following line if your bare application
	// has an action associated with it:
}

func Execute(args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) int {
	project := internal.NewProject()
	currentDir, err := os.Getwd()
	if err != nil {
		internal.Red.Fprintln(stderr, "Failed to execute command: %v", err)
		return -1
	}
	if project.Root() != currentDir {
		internal.Yellow.Fprintln(stderr, "Please execute the command in the project root dir")
		return -1
	}
	ctx := context.Background()
	ctx = context.WithValue(ctx, proCtx, project)
	if err = rootCmd.ExecuteContext(ctx); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			return exitError.ExitCode()
		} else {
			return 1
		}
	}
	return 0
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.gob.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.SetErrPrefix(internal.Red.Sprintf("Error:"))
	rootCmd.AddCommand(initCmd)
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
