package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
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

func detectShell(args []string) (string, error) {
	if len(args) == 1 {
		shell := normalizeShell(args[0])
		if shell == "" {
			return "", fmt.Errorf("unsupported shell %q (expected bash or zsh)", args[0])
		}
		return shell, nil
	}

	if shell := detectParentShell(); shell != "" {
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

type shellSetupResult struct {
	PrimaryPath      string
	ExtraPath        string
	AlreadyInstalled bool
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

func detectParentShell() string {
	ppid := strconv.Itoa(os.Getppid())
	cmd := exec.Command("ps", "-p", ppid, "-o", "comm=")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return ""
	}

	return normalizeShell(strings.TrimSpace(out.String()))
}

func installShellSetup(homeDir, shell string) (shellSetupResult, error) {
	primaryPath := shellRCPath(homeDir, shell)
	alreadyInstalled, err := installShellSetupLine(primaryPath, shellSetupLine(shell), "# awry shell integration")
	if err != nil {
		return shellSetupResult{}, err
	}

	result := shellSetupResult{
		PrimaryPath:      primaryPath,
		AlreadyInstalled: alreadyInstalled,
	}

	if shell == "bash" {
		profilePath := filepath.Join(homeDir, ".bash_profile")
		profileLine := "if [ -f ~/.bashrc ]; then . ~/.bashrc; fi"
		profileInstalled, err := installShellSetupLine(profilePath, profileLine, "# awry bashrc loader")
		if err != nil {
			return shellSetupResult{}, err
		}
		if !profileInstalled {
			result.ExtraPath = profilePath
		}
	}

	return result, nil
}

func installShellSetupLine(rcPath, line, comment string) (bool, error) {
	content, err := os.ReadFile(rcPath)
	if err != nil && !os.IsNotExist(err) {
		return false, fmt.Errorf("reading %s: %w", rcPath, err)
	}

	text := string(content)
	if strings.Contains(text, line) {
		return true, nil
	}

	block := "\n" + comment + "\n" + line + "\n"
	if text == "" {
		block = comment + "\n" + line + "\n"
	}

	updated := text + block
	if err := os.WriteFile(rcPath, []byte(updated), 0o644); err != nil {
		return false, fmt.Errorf("writing %s: %w", rcPath, err)
	}

	return false, nil
}
