package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init [bash|zsh]",
	Short: "Print shell wrapper setup for bash or zsh",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		shell, err := detectShell(args)
		if err != nil {
			return err
		}

		fmt.Print(shellInitScript(shell))
		return nil
	},
}

var setupShellCmd = &cobra.Command{
	Use:   "setup-shell [bash|zsh]",
	Short: "Install shell setup so `awry` updates your current shell",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		shell, err := detectShell(args)
		if err != nil {
			return err
		}

		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("detecting home directory: %w", err)
		}

		result, err := installShellSetup(homeDir, shell)
		if err != nil {
			return err
		}

		if result.AlreadyInstalled {
			fmt.Printf("awry shell setup already exists in %s\n", result.PrimaryPath)
		} else {
			fmt.Printf("Added awry shell setup to %s\n", result.PrimaryPath)
		}

		if result.ExtraPath != "" {
			fmt.Printf("Ensured %s loads %s\n", result.ExtraPath, result.PrimaryPath)
		}

		fmt.Printf("Run: source %s\n", result.PrimaryPath)
		return nil
	},
}
