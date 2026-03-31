package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

func detectShell(args []string) (string, error) {
	if len(args) == 1 {
		shell := normalizeShell(args[0])
		if shell == "" {
			return "", fmt.Errorf("unsupported shell %q (expected bash, zsh, or fish)", args[0])
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

	return "", fmt.Errorf("could not detect shell, pass one explicitly: awry init bash, awry init zsh, or awry init fish")
}

func normalizeShell(shell string) string {
	if shell == "" {
		return ""
	}

	base := strings.ToLower(filepath.Base(shell))
	switch base {
	case "bash", "zsh", "fish":
		return base
	default:
		return ""
	}
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
