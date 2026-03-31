package app

import (
	"testing"

	"github.com/Art-Thor/awry/pkg/shellenv"
)

func TestSelectionHintCommand(t *testing.T) {
	got := selectionHintCommand("team's sandbox", "")
	want := `eval "export AWS_PROFILE='team'\"'\"'s sandbox'"`
	if got != want {
		t.Fatalf("selectionHintCommand() = %q, want %q", got, want)
	}
}

func TestSelectionHintCommandFish(t *testing.T) {
	got := selectionHintCommand("team's sandbox", shellenv.ShellFish)
	want := `set -gx AWS_PROFILE 'team\'s sandbox' | source`
	if got != want {
		t.Fatalf("selectionHintCommand() = %q, want %q", got, want)
	}
}
