package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNormalizeShell(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "bash name", input: "bash", want: "bash"},
		{name: "zsh path", input: "/bin/zsh", want: "zsh"},
		{name: "unsupported", input: "fish", want: ""},
		{name: "empty", input: "", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeShell(tt.input); got != tt.want {
				t.Fatalf("normalizeShell(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestDetectShell(t *testing.T) {
	origShell := os.Getenv("SHELL")
	origAwryShell := os.Getenv("AWRY_SHELL")
	t.Cleanup(func() {
		_ = os.Setenv("SHELL", origShell)
		_ = os.Setenv("AWRY_SHELL", origAwryShell)
	})

	t.Run("explicit arg wins", func(t *testing.T) {
		shell, err := detectShell([]string{"zsh"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if shell != "zsh" {
			t.Fatalf("expected zsh, got %q", shell)
		}
	})

	t.Run("env fallback", func(t *testing.T) {
		_ = os.Setenv("AWRY_SHELL", "")
		_ = os.Setenv("SHELL", "/bin/bash")

		shell, err := detectShell(nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if shell != "bash" {
			t.Fatalf("expected bash, got %q", shell)
		}
	})

	t.Run("unsupported shell errors", func(t *testing.T) {
		_ = os.Setenv("AWRY_SHELL", "fish")
		_ = os.Setenv("SHELL", "fish")

		if _, err := detectShell(nil); err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestShellInitScript(t *testing.T) {
	script := shellInitScript("zsh")

	checks := []string{
		"awry() {",
		`command awry "$@"`,
		`eval "$_awry_output"`,
		`""|use)`,
	}

	for _, check := range checks {
		if !strings.Contains(script, check) {
			t.Fatalf("expected script to contain %q", check)
		}
	}
}

func TestShellRCPath(t *testing.T) {
	if got := shellRCPath("/tmp/home", "bash"); got != "/tmp/home/.bashrc" {
		t.Fatalf("unexpected bash rc path: %q", got)
	}

	if got := shellRCPath("/tmp/home", "zsh"); got != "/tmp/home/.zshrc" {
		t.Fatalf("unexpected zsh rc path: %q", got)
	}
}

func TestShellSetupLine(t *testing.T) {
	if got := shellSetupLine("zsh"); got != `eval "$(command awry init zsh)"` {
		t.Fatalf("unexpected setup line: %q", got)
	}
}

func TestInstallShellSetup(t *testing.T) {
	t.Run("creates zsh config entry", func(t *testing.T) {
		homeDir := t.TempDir()
		rcPath := filepath.Join(homeDir, ".zshrc")

		result, err := installShellSetup(homeDir, "zsh")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.AlreadyInstalled {
			t.Fatal("expected fresh install")
		}
		if result.PrimaryPath != rcPath {
			t.Fatalf("unexpected primary path: %q", result.PrimaryPath)
		}

		content, err := os.ReadFile(rcPath)
		if err != nil {
			t.Fatalf("reading rc file: %v", err)
		}

		text := string(content)
		if !strings.Contains(text, "# awry shell integration") {
			t.Fatalf("expected awry comment in %q", text)
		}
		if !strings.Contains(text, shellSetupLine("zsh")) {
			t.Fatalf("expected setup line in %q", text)
		}
	})

	t.Run("creates bash config and loader", func(t *testing.T) {
		homeDir := t.TempDir()
		rcPath := filepath.Join(homeDir, ".bashrc")
		profilePath := filepath.Join(homeDir, ".bash_profile")

		result, err := installShellSetup(homeDir, "bash")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.PrimaryPath != rcPath {
			t.Fatalf("unexpected primary path: %q", result.PrimaryPath)
		}

		profileContent, err := os.ReadFile(profilePath)
		if err != nil {
			t.Fatalf("reading bash profile: %v", err)
		}
		if !strings.Contains(string(profileContent), "if [ -f ~/.bashrc ]; then . ~/.bashrc; fi") {
			t.Fatalf("expected bash profile loader in %q", string(profileContent))
		}
	})

	t.Run("does not duplicate config entry", func(t *testing.T) {
		homeDir := t.TempDir()
		rcPath := filepath.Join(homeDir, ".bashrc")
		if err := os.WriteFile(rcPath, []byte(shellSetupLine("bash")+"\n"), 0o644); err != nil {
			t.Fatalf("writing rc file: %v", err)
		}

		result, err := installShellSetup(homeDir, "bash")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.AlreadyInstalled {
			t.Fatal("expected existing install to be detected")
		}
	})
}
