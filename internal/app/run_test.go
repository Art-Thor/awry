package app

import "testing"

func TestSelectionHintCommand(t *testing.T) {
	got := selectionHintCommand("team's sandbox")
	want := `eval "export AWS_PROFILE='team'\"'\"'s sandbox'"`
	if got != want {
		t.Fatalf("selectionHintCommand() = %q, want %q", got, want)
	}
}
