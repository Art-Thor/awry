package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

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
