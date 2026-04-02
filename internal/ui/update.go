package ui

import (
	"context"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Art-Thor/awry/internal/identity"
	"github.com/Art-Thor/awry/internal/session"
	"github.com/Art-Thor/awry/pkg/models"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case sessionLoadedMsg:
		m.sessionInfo = msg.info
		m.sessionErr = msg.err
		return m, nil
	case identityLoadedMsg:
		m.identity = msg.identity
		m.identityErr = msg.err
		return m, nil
	}
	return m, nil
}

type sessionLoadedMsg struct {
	info *session.Info
	err  error
}

type identityLoadedMsg struct {
	identity *identity.Identity
	err      error
}

func loadSessionCmd(profile models.Profile) tea.Cmd {
	return func() tea.Msg {
		info, err := resolveSession(profile)
		if err != nil {
			return sessionLoadedMsg{err: err}
		}
		return sessionLoadedMsg{info: &info}
	}
}

func loadIdentityCmd(profile string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		resolved, err := lookupIdentity(ctx, profile)
		if err != nil {
			return identityLoadedMsg{err: err}
		}
		return identityLoadedMsg{identity: &resolved}
	}
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.helpVisible {
		switch msg.String() {
		case "?", "esc", "q":
			m.helpVisible = false
		}
		return m, nil
	}

	if m.searching {
		switch msg.Type {
		case tea.KeyEsc:
			m.searching = false
			m.searchQuery = ""
			m.filtered = m.profiles
			m.cursor = 0
			return m, nil
		case tea.KeyEnter:
			m.searching = false
			return m, nil
		case tea.KeyBackspace:
			if len(m.searchQuery) > 0 {
				m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
				m.applyFilter()
			}
			return m, nil
		default:
			if msg.Type == tea.KeyRunes {
				m.searchQuery += string(msg.Runes)
				m.applyFilter()
			}
			return m, nil
		}
	}

	switch msg.String() {
	case "q", "ctrl+c":
		m.quitting = true
		return m, tea.Quit
	case "?":
		m.helpVisible = true
		return m, nil
	case "r":
		m.sessionInfo = nil
		m.sessionErr = nil
		m.identity = nil
		m.identityErr = nil
		return m, m.refreshActiveRuntimeCmd()
	case "p":
		if len(m.filtered) == 0 {
			return m, nil
		}
		if err := m.toggleFavorite(m.filtered[m.cursor].Name); err != nil {
			return m, nil
		}
		return m, nil
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.filtered)-1 {
			m.cursor++
		}
	case "/":
		m.searching = true
		m.searchQuery = ""
	case "enter":
		if len(m.filtered) > 0 {
			p := m.filtered[m.cursor]
			_ = m.recordRecent(p.Name)
			m.selected = &p
			return m, tea.Quit
		}
	}

	return m, nil
}
