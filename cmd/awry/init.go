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
	Short: "Print shell setup for automatic profile switching",
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
