package app

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Art-Thor/awry/internal/ui"
	"github.com/Art-Thor/awry/pkg/shellenv"
)

// RunTUI launches the interactive profile picker and prints the export
// command for the selected profile.
func RunTUI() error {
	m, err := ui.New()
	if err != nil {
		return err
	}

	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithOutput(os.Stderr))
	result, err := p.Run()
	if err != nil {
		return fmt.Errorf("running TUI: %w", err)
	}

	final := result.(ui.Model)
	if sel := final.SelectedProfile(); sel != nil {
		shell := os.Getenv("AWRY_SHELL")
		export := ExportCommandForShell(sel.Name, shell)
		if stdoutIsTerminal() {
			fmt.Fprintf(os.Stderr, "\nSelected: %s\nTo let `awry` update your shell automatically, run once:\n  awry setup-shell\nFor this selection only, run:\n  %s\n\n", sel.Name, selectionHintCommand(sel.Name, shell))
		} else {
			fmt.Fprintf(os.Stderr, "\nSelected: %s\n\n", sel.Name)
		}
		fmt.Println(export)
	}

	return nil
}

func stdoutIsTerminal() bool {
	info, err := os.Stdout.Stat()
	if err != nil {
		return false
	}

	return info.Mode()&os.ModeCharDevice != 0
}

func selectionHintCommand(profile, shell string) string {
	if shell == shellenv.ShellFish {
		return fmt.Sprintf("%s | source", ExportCommandForShell(profile, shell))
	}

	return fmt.Sprintf("eval %q", ExportCommandForShell(profile, shell))
}
