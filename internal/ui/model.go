package ui

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/Art-Thor/awry/internal/awsconfig"
	"github.com/Art-Thor/awry/internal/identity"
	"github.com/Art-Thor/awry/internal/session"
	"github.com/Art-Thor/awry/pkg/models"
)

var (
	lookupIdentity = identity.Lookup
	resolveSession = session.ResolveProfile
)

// Model is the top-level Bubble Tea model.
type Model struct {
	profiles       []models.Profile
	filtered       []models.Profile
	cursor         int
	currentProfile string
	searchQuery    string
	searching      bool
	selected       *models.Profile
	width          int
	height         int
	quitting       bool
	identity       *identity.Identity
	identityErr    error
	sessionInfo    *session.Info
	sessionErr     error
}

// SelectedProfile returns the profile the user chose (nil if none).
func (m Model) SelectedProfile() *models.Profile {
	return m.selected
}

// New creates the initial model.
func New() (Model, error) {
	profiles, err := awsconfig.LoadProfiles()
	if err != nil {
		return Model{}, fmt.Errorf("loading profiles: %w", err)
	}

	m := Model{
		profiles:       profiles,
		currentProfile: awsconfig.CurrentProfile(),
	}
	m.pinActiveToTop()
	m.filtered = m.profiles
	return m, nil
}

func (m Model) Init() tea.Cmd {
	active, ok := m.activeProfile()
	if !ok {
		return nil
	}

	return tea.Batch(loadSessionCmd(active), loadIdentityCmd(active.Name))
}

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
	// When in search mode, handle text input.
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
			m.selected = &p
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m *Model) applyFilter() {
	if m.searchQuery == "" {
		m.filtered = m.profiles
		m.cursor = 0
		return
	}

	query := strings.ToLower(m.searchQuery)
	var result []models.Profile
	for _, p := range m.profiles {
		if fuzzyMatch(strings.ToLower(p.Name), query) {
			result = append(result, p)
		}
	}
	m.filtered = result
	m.cursor = 0
}

// pinActiveToTop moves the active profile to the first position in the list.
func (m *Model) pinActiveToTop() {
	if m.currentProfile == "" {
		return
	}
	for i, p := range m.profiles {
		if p.Name == m.currentProfile && i > 0 {
			m.profiles = append([]models.Profile{p}, append(m.profiles[:i], m.profiles[i+1:]...)...)
			return
		}
	}
}

func (m Model) activeProfile() (models.Profile, bool) {
	if m.currentProfile == "" {
		return models.Profile{}, false
	}
	for _, profile := range m.profiles {
		if profile.Name == m.currentProfile {
			return profile, true
		}
	}
	return models.Profile{}, false
}

// inlineBadge returns the inline type badge string for a profile.
func inlineBadge(t models.ProfileType) string {
	switch t {
	case models.ProfileTypeSSO:
		return badgeInlineSSO.String()
	case models.ProfileTypeRole:
		return badgeInlineRole.String()
	case models.ProfileTypeStatic:
		return badgeInlineStatic.String()
	default:
		return badgeInlineUnknown.String()
	}
}

// fuzzyMatch checks if all characters in query appear in order within s.
func fuzzyMatch(s, query string) bool {
	qi := 0
	for i := 0; i < len(s) && qi < len(query); i++ {
		if s[i] == query[qi] {
			qi++
		}
	}
	return qi == len(query)
}

func (m Model) View() string {
	if m.quitting {
		return ""
	}

	if m.width == 0 {
		return "Loading..."
	}

	listWidth := m.width * 2 / 5
	if listWidth < 30 {
		listWidth = 30
	}
	detailWidth := m.width - listWidth - 4 // border + padding

	list := m.renderList(listWidth)
	detail := m.renderDetail(detailWidth)
	status := m.renderStatusBar()

	panels := lipgloss.JoinHorizontal(
		lipgloss.Top,
		listPanelStyle.Width(listWidth).Render(list),
		detailPanelStyle.Width(detailWidth).Render(detail),
	)

	return lipgloss.JoinVertical(lipgloss.Left, panels, status)
}

func (m Model) renderList(width int) string {
	var b strings.Builder

	title := titleStyle.Render("AWS Profiles")
	b.WriteString(title)
	b.WriteString("\n\n")

	if m.searching {
		b.WriteString(searchStyle.Render("/ " + m.searchQuery + "█"))
		b.WriteString("\n\n")
	}

	if len(m.filtered) == 0 {
		b.WriteString(lipgloss.NewStyle().Foreground(colorSecondary).Render("No profiles found"))
		return b.String()
	}

	// Calculate visible window for scrolling.
	maxVisible := m.height - 8
	if m.searching {
		maxVisible -= 2
	}
	if maxVisible < 3 {
		maxVisible = 3
	}

	start := 0
	if m.cursor >= maxVisible {
		start = m.cursor - maxVisible + 1
	}
	end := start + maxVisible
	if end > len(m.filtered) {
		end = len(m.filtered)
	}

	for i := start; i < end; i++ {
		p := m.filtered[i]
		name := p.Name + inlineBadge(p.Type)

		isActive := p.Name == m.currentProfile
		if isActive {
			name += " ●"
		}

		var line string
		if i == m.cursor {
			line = selectedItemStyle.Render(name)
		} else if isActive {
			line = normalItemStyle.Render(activeProfileStyle.Render(name))
		} else {
			line = normalItemStyle.Render(name)
		}

		b.WriteString(line)
		b.WriteString("\n")

		// Divider after the active profile (when pinned to top).
		if isActive && i == 0 && len(m.filtered) > 1 && !m.searching {
			b.WriteString(dividerStyle.Render("  ─────────────────────"))
			b.WriteString("\n")
		}
	}

	return b.String()
}

func (m Model) renderDetail(width int) string {
	if len(m.filtered) == 0 {
		return lipgloss.NewStyle().Foreground(colorSecondary).Render("No profile selected")
	}

	p := m.filtered[m.cursor]

	var b strings.Builder

	b.WriteString(detailTitleStyle.Render(p.Name))
	b.WriteString("\n\n")

	badge := badgeFor(p.Type)
	b.WriteString(row("Type", badge))
	b.WriteString(row("Region", p.DisplayRegion()))

	if p.SourceProfile != "" {
		b.WriteString(row("Source Profile", p.SourceProfile))
	}
	if p.RoleARN != "" {
		b.WriteString(row("Role ARN", p.RoleARN))
	}
	if p.SSOStartURL != "" {
		b.WriteString(row("SSO URL", p.SSOStartURL))
	}
	if p.SSOAccountID != "" {
		b.WriteString(row("SSO Account", p.SSOAccountID))
	}
	if p.SSORegion != "" {
		b.WriteString(row("SSO Region", p.SSORegion))
	}
	if p.SSORoleName != "" {
		b.WriteString(row("SSO Role", p.SSORoleName))
	}
	if p.Output != "" {
		b.WriteString(row("Output", p.Output))
	}

	if p.Name == m.currentProfile {
		b.WriteString(m.renderActiveRuntimeDetails())
		b.WriteString("\n")
		b.WriteString(activeProfileStyle.Render("● Currently active"))
	}

	return b.String()
}

func (m Model) renderActiveRuntimeDetails() string {
	var b strings.Builder

	b.WriteString(row("Session", m.sessionStatusValue()))

	if m.identity != nil {
		b.WriteString(row("Account ID", m.identity.AccountID))
		b.WriteString(row("ARN", m.identity.ARN))
		b.WriteString(row("Principal", m.identity.Principal))
	} else if m.identityErr != nil {
		b.WriteString(row("Identity", m.identityErr.Error()))
	} else {
		b.WriteString(row("Identity", "Loading..."))
	}

	return b.String()
}

func (m Model) sessionStatusValue() string {
	if m.sessionErr != nil {
		return m.sessionErr.Error()
	}
	if m.sessionInfo == nil {
		return "Loading..."
	}

	switch m.sessionInfo.Status {
	case session.StatusNotApplicable:
		return "No expiry"
	case session.StatusUnknown:
		return "Unknown"
	case session.StatusExpired:
		return "Expired"
	case session.StatusExpiringSoon:
		return fmt.Sprintf("%s left (expiring soon)", formatDuration(m.sessionInfo.Remaining))
	case session.StatusActive:
		return fmt.Sprintf("%s left", formatDuration(m.sessionInfo.Remaining))
	default:
		return string(m.sessionInfo.Status)
	}
}

func formatDuration(d time.Duration) string {
	if d <= 0 {
		return "0m"
	}
	hours := int(d / time.Hour)
	minutes := int((d % time.Hour) / time.Minute)
	if hours == 0 {
		return fmt.Sprintf("%dm", minutes)
	}
	if minutes == 0 {
		return fmt.Sprintf("%dh", hours)
	}
	return fmt.Sprintf("%dh %dm", hours, minutes)
}

func row(label, value string) string {
	return detailLabelStyle.Render(label) + detailValueStyle.Render(value) + "\n"
}

func badgeFor(t models.ProfileType) string {
	switch t {
	case models.ProfileTypeSSO:
		return badgeSSO.String()
	case models.ProfileTypeRole:
		return badgeRole.String()
	case models.ProfileTypeStatic:
		return badgeStatic.String()
	default:
		return badgeUnknown.String()
	}
}

func (m Model) renderStatusBar() string {
	var parts []string
	if m.searching {
		parts = append(parts, "Esc close search", "Enter confirm")
	} else {
		parts = append(parts, "↑↓/jk navigate", "Enter select profile", "/ search", "q quit")
	}
	return statusBarStyle.Render(strings.Join(parts, "  │  "))
}
