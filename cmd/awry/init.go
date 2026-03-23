package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

		rcPath := shellRCPath(homeDir, shell)
		alreadyInstalled, err := installShellSetup(rcPath, shell)
		if err != nil {
			return err
		}

		if alreadyInstalled {
			fmt.Printf("awry shell setup already exists in %s\n", rcPath)
		} else {
			fmt.Printf("Added awry shell setup to %s\n", rcPath)
		}

		fmt.Printf("Run: source %s\n", rcPath)
		return nil
	},
}

func detectShell(args []string) (string, error) {
	if len(args) == 1 {
		shell := normalizeShell(args[0])
		if shell == "" {
			return "", fmt.Errorf("unsupported shell %q (expected bash or zsh)", args[0])
		}
		return shell, nil
	}

	if shell := normalizeShell(os.Getenv("AWRY_SHELL")); shell != "" {
		return shell, nil
	}

	if shell := normalizeShell(os.Getenv("SHELL")); shell != "" {
		return shell, nil
	}

	return "", fmt.Errorf("could not detect shell, pass one explicitly: awry init bash or awry init zsh")
}

func normalizeShell(shell string) string {
	if shell == "" {
		return ""
	}

	base := strings.ToLower(filepath.Base(shell))
	switch base {
	case "bash", "zsh":
		return base
	default:
		return ""
	}
}

func shellInitScript(shell string) string {
	_ = shell
	return `awry() {
  case "$1" in
    ""|use)
      local _awry_output
      _awry_output="$(command awry "$@")" || return $?
      if [ -n "$_awry_output" ]; then
        eval "$_awry_output"
      fi
      ;;
    *)
      command awry "$@"
      ;;
  esac
}
`
}

func shellRCPath(homeDir, shell string) string {
	fileName := ".bashrc"
	if shell == "zsh" {
		fileName = ".zshrc"
	}

	return filepath.Join(homeDir, fileName)
}

func shellSetupLine(shell string) string {
	return fmt.Sprintf(`eval "$(command awry init %s)"`, shell)
}

func installShellSetup(rcPath, shell string) (bool, error) {
	content, err := os.ReadFile(rcPath)
	if err != nil && !os.IsNotExist(err) {
		return false, fmt.Errorf("reading %s: %w", rcPath, err)
	}

	line := shellSetupLine(shell)
	text := string(content)
	if strings.Contains(text, line) {
		return true, nil
	}

	block := "\n# awry shell integration\n" + line + "\n"
	if text == "" {
		block = "# awry shell integration\n" + line + "\n"
	}

	updated := text + block
	if err := os.WriteFile(rcPath, []byte(updated), 0o644); err != nil {
		return false, fmt.Errorf("writing %s: %w", rcPath, err)
	}

	return false, nil
}
