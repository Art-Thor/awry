package app

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Art-Thor/awry/internal/ui"
)

// RunTUI launches the interactive profile picker and prints the export
// command for the selected profile.
func RunTUI() error {
	m, err := ui.New()
	if err != nil {
		return err
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	result, err := p.Run()
	if err != nil {
		return fmt.Errorf("running TUI: %w", err)
	}

	final := result.(ui.Model)
	if sel := final.SelectedProfile(); sel != nil {
		fmt.Fprintf(os.Stderr, "\n  Selected: %s\n\n", sel.Name)
		fmt.Printf("export AWS_PROFILE=%s\n", sel.Name)
	}

	return nil
}
