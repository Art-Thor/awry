package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/Art-Thor/awry/pkg/models"
	"github.com/Art-Thor/awry/pkg/shellenv"
	"os"
)

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
	detailWidth := m.width - listWidth - 4

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

	b.WriteString(titleStyle.Render("AWS Profiles"))
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
		if header := m.sectionHeader(i); header != "" {
			b.WriteString(sectionHeaderStyle.Render(header))
			b.WriteString("\n")
		}

		p := m.filtered[i]
		name := riskMarker(m.isRiskyProfile(p.Name)) + recentMarker(m.isRecent(p.Name)) + favoriteMarker(m.isFavorite(p.Name)) + p.Name + inlineBadge(p.Type) + m.listHealthBadge(p)

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

		if isActive && i == 0 && len(m.filtered) > 1 && !m.searching {
			b.WriteString(dividerStyle.Render("  ─────────────────────"))
			b.WriteString("\n")
		}
	}

	return b.String()
}

func (m Model) sectionHeader(index int) string {
	if m.searching || index < 0 || index >= len(m.filtered) {
		return ""
	}

	current := m.filtered[index]
	if index == 0 {
		switch {
		case current.Name == m.currentProfile:
			return "Current"
		case m.isFavorite(current.Name):
			return "Favorites"
		case m.isRecent(current.Name):
			return "Recent"
		default:
			return "All Profiles"
		}
	}

	previous := m.filtered[index-1]
	if current.Name == m.currentProfile {
		return ""
	}
	if m.isFavorite(current.Name) && !m.isFavorite(previous.Name) {
		return "Favorites"
	}
	if m.isRecent(current.Name) && !m.isRecent(previous.Name) && !m.isFavorite(current.Name) {
		return "Recent"
	}
	if !m.isFavorite(current.Name) && !m.isRecent(current.Name) && (m.isFavorite(previous.Name) || m.isRecent(previous.Name) || previous.Name == m.currentProfile) {
		return "All Profiles"
	}

	return ""
}

func (m Model) renderDetail(width int) string {
	if len(m.filtered) == 0 {
		return lipgloss.NewStyle().Foreground(colorSecondary).Render("No profile selected")
	}

	p := m.filtered[m.cursor]

	var b strings.Builder

	b.WriteString(detailTitleStyle.Render(p.Name))
	b.WriteString("\n\n")
	b.WriteString(row("Favorite", yesNo(m.isFavorite(p.Name))))
	b.WriteString(row("Recent", yesNo(m.isRecent(p.Name))))
	b.WriteString(row("Risk", m.riskLabel(p.Name)))
	b.WriteString(row("Type", badgeFor(p.Type)))
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

func (m Model) renderHelpOverlay() string {
	rows := []string{
		helpTitleStyle.Render("Keyboard Help"),
		"",
		helpKeyStyle.Render("j / k, up / down") + helpDescStyle.Render("Move through profiles"),
		helpKeyStyle.Render("Enter") + helpDescStyle.Render("Select the highlighted profile"),
		helpKeyStyle.Render("p") + helpDescStyle.Render("Toggle favorite on the highlighted profile"),
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
	return shellenv.ExportCommandForShell(profile, os.Getenv("AWRY_SHELL"))
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
		parts = append(parts, "↑↓/jk navigate", "Enter select profile", "p favorite", "r refresh", "? help", "/ search", "q quit")
	}
	return statusBarStyle.Render(strings.Join(parts, "  │  "))
}

func recentMarker(isRecent bool) string {
	if isRecent {
		return "> "
	}
	return "  "
}

func riskMarker(isRisky bool) string {
	if isRisky {
		return "! "
	}
	return "  "
}

func favoriteMarker(isFavorite bool) string {
	if isFavorite {
		return "* "
	}
	return "  "
}

func yesNo(v bool) string {
	if v {
		return "Yes"
	}
	return "No"
}
