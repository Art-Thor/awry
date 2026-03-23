package main

import (
	"os"
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
