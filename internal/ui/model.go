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
	helpVisible    bool
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
	return m.refreshActiveRuntimeCmd()
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
	if m.helpVisible {
		switch msg.String() {
		case "?", "esc", "q":
			m.helpVisible = false
		}
		return m, nil
	}

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
	case "?":
		m.helpVisible = true
		return m, nil
	case "r":
		m.sessionInfo = nil
		m.sessionErr = nil
		m.identity = nil
		m.identityErr = nil
		return m, m.refreshActiveRuntimeCmd()
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

func (m Model) refreshActiveRuntimeCmd() tea.Cmd {
	active, ok := m.activeProfile()
	if !ok {
		return nil
	}

	return tea.Batch(loadSessionCmd(active), loadIdentityCmd(active.Name))
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

	view := lipgloss.JoinVertical(lipgloss.Left, panels, status)
	if m.helpVisible {
		return view + "\n\n" + m.renderHelpOverlay()
	}

	return view
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
		if m.searchQuery != "" {
			b.WriteString(lipgloss.NewStyle().Foreground(colorSecondary).Render("No profiles match search"))
			b.WriteString("\n")
			b.WriteString(lipgloss.NewStyle().Foreground(colorMuted).Render("Press Esc to clear the filter"))
			return b.String()
		}

		b.WriteString(lipgloss.NewStyle().Foreground(colorSecondary).Render("No AWS profiles found"))
		b.WriteString("\n")
		b.WriteString(lipgloss.NewStyle().Foreground(colorMuted).Render("Check ~/.aws/config and ~/.aws/credentials"))
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
		name := p.Name + inlineBadge(p.Type) + m.listHealthBadge(p)

		isActive := p.Name == m.currentProfile
		if isActive {
			name += " ●" + m.activeRuntimeBadge()
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
	b.WriteString(row("Health", m.profileHealthValue(p)))
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
		b.WriteString(m.renderActiveRuntimeDetails(p))
		b.WriteString("\n")
		b.WriteString(activeProfileStyle.Render("● Currently active"))
	} else {
		b.WriteString(row("Export", exportCommand(p.Name)))
	}

	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Foreground(colorMuted).Render("Press Enter to emit the export command for this profile."))

	return b.String()
}

func (m Model) renderActiveRuntimeDetails(profile models.Profile) string {
	var b strings.Builder

	b.WriteString(row("Session", m.sessionStatusValue(profile)))

	if m.identity != nil {
		b.WriteString(row("Account ID", m.identity.AccountID))
		b.WriteString(row("ARN", m.identity.ARN))
		b.WriteString(row("Principal", m.identity.Principal))
	} else if m.identityErr != nil {
		b.WriteString(row("Identity", identityStatusValue(m.identityErr)))
	} else {
		b.WriteString(row("Identity", "Loading..."))
	}

	return b.String()
}

func (m Model) sessionStatusValue(profile models.Profile) string {
	if m.sessionErr != nil {
		return "Unavailable"
	}
	if m.sessionInfo == nil {
		return "Loading..."
	}

	switch m.sessionInfo.Status {
	case session.StatusNotApplicable:
		if profile.Type == models.ProfileTypeStatic {
			return "No expiry (static credentials)"
		}
		return "No expiry"
	case session.StatusUnknown:
		if profile.Type == models.ProfileTypeSSO {
			return "Unknown - run aws sso login if needed"
		}
		return "Unknown"
	case session.StatusExpired:
		return "Expired - refresh credentials"
	case session.StatusExpiringSoon:
		return fmt.Sprintf("%s left (expiring soon)", formatDuration(m.sessionInfo.Remaining))
	case session.StatusActive:
		return fmt.Sprintf("%s left", formatDuration(m.sessionInfo.Remaining))
	default:
		return string(m.sessionInfo.Status)
	}
}

func (m Model) activeRuntimeBadge() string {
	if m.sessionErr != nil || m.identityErr != nil {
		if isNoCredentialsError(m.identityErr) {
			return runtimeBadgeNoCreds.String()
		}
		if isInvalidIdentityError(m.identityErr) {
			return runtimeBadgeInvalid.String()
		}
		return runtimeBadgeError.String()
	}
	if m.sessionInfo == nil {
		return runtimeBadgeLoading.String()
	}

	switch m.sessionInfo.Status {
	case session.StatusExpired:
		return runtimeBadgeExpired.String()
	case session.StatusExpiringSoon:
		return runtimeBadgeExpiring.String()
	case session.StatusActive:
		return runtimeBadgeOK.String()
	case session.StatusNotApplicable:
		return runtimeBadgeOK.String()
	case session.StatusUnknown:
		return runtimeBadgeUnknown.String()
	default:
		return runtimeBadgeInfo.String()
	}
}

func (m Model) listHealthBadge(profile models.Profile) string {
	if profile.Name != m.currentProfile {
		if profile.Type == models.ProfileTypeUnknown {
			return runtimeBadgeInvalid.String()
		}
		if profile.Type == models.ProfileTypeRole && profile.SourceProfile == "" {
			return runtimeBadgeInvalid.String()
		}
		return ""
	}

	return m.activeRuntimeBadge()
}

func (m Model) profileHealthValue(profile models.Profile) string {
	if profile.Type == models.ProfileTypeUnknown {
		return "Invalid - profile is missing auth configuration"
	}

	if profile.Type == models.ProfileTypeRole && profile.SourceProfile == "" {
		return "Invalid - role profile has no source_profile"
	}

	if profile.Name != m.currentProfile {
		switch profile.Type {
		case models.ProfileTypeSSO:
			return "Needs runtime check when active"
		case models.ProfileTypeStatic:
			if profile.HasCredentials {
				return "Configured"
			}
			return "Invalid - static profile has no credentials"
		default:
			return "Configured"
		}
	}

	if m.sessionErr != nil {
		return "Unavailable - unable to inspect session"
	}
	if m.identityErr != nil {
		if isNoCredentialsError(m.identityErr) {
			return "No credentials available"
		}
		if isInvalidIdentityError(m.identityErr) {
			return "Invalid - AWS configuration rejected"
		}
		return "Check configuration"
	}

	return m.sessionStatusValue(profile)
}

func identityStatusValue(err error) string {
	if err == nil {
		return "Loading..."
	}

	message := err.Error()
	switch {
	case strings.Contains(message, "expired"):
		return "Expired - run aws sso login"
	case isNoCredentialsError(err):
		return "No credentials available"
	default:
		return "Unavailable"
	}
}

func isNoCredentialsError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(strings.ToLower(err.Error()), "no valid aws credentials")
}

func isInvalidIdentityError(err error) bool {
	if err == nil {
		return false
	}

	message := strings.ToLower(err.Error())
	return strings.Contains(message, "could not load config") ||
		strings.Contains(message, "invalid configuration") ||
		strings.Contains(message, "partial credentials")
}

func (m Model) renderHelpOverlay() string {
	rows := []string{
		helpTitleStyle.Render("Keyboard Help"),
		"",
		helpKeyStyle.Render("j / k, up / down") + helpDescStyle.Render("Move through profiles"),
		helpKeyStyle.Render("Enter") + helpDescStyle.Render("Select the highlighted profile"),
		helpKeyStyle.Render("/") + helpDescStyle.Render("Start fuzzy search"),
		helpKeyStyle.Render("r") + helpDescStyle.Render("Refresh active session and identity"),
		helpKeyStyle.Render("?") + helpDescStyle.Render("Toggle this help overlay"),
		helpKeyStyle.Render("q") + helpDescStyle.Render("Quit awry"),
		"",
		helpFooterStyle.Render("Press ? or Esc to close"),
	}

	return helpBoxStyle.Render(strings.Join(rows, "\n"))
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

func exportCommand(profile string) string {
	return fmt.Sprintf("export AWS_PROFILE=%s", shellQuote(profile))
}

func shellQuote(s string) string {
	if s == "" {
		return "''"
	}

	quoted := "'"
	for _, r := range s {
		if r == '\'' {
			quoted += `"'"'`
			continue
		}
		quoted += string(r)
	}
	quoted += "'"

	return quoted
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
		parts = append(parts, "↑↓/jk navigate", "Enter select profile", "r refresh", "? help", "/ search", "q quit")
	}
	return statusBarStyle.Render(strings.Join(parts, "  │  "))
}
