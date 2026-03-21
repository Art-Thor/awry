package ui

import "github.com/charmbracelet/lipgloss"

var (
	// Colors
	colorPrimary   = lipgloss.Color("#7C3AED") // violet
	colorSecondary = lipgloss.Color("#6B7280") // gray
	colorSuccess   = lipgloss.Color("#10B981") // green
	colorWarning   = lipgloss.Color("#F59E0B") // amber
	colorInfo      = lipgloss.Color("#3B82F6") // blue
	colorMuted     = lipgloss.Color("#4B5563")
	colorBg        = lipgloss.Color("#1F2937")
	colorWhite     = lipgloss.Color("#F9FAFB")

	// Panel styles
	listPanelStyle = lipgloss.NewStyle().
			Padding(1, 2)

	detailPanelStyle = lipgloss.NewStyle().
				Padding(1, 2).
				BorderLeft(true).
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(colorMuted)

	// Item styles
	normalItemStyle = lipgloss.NewStyle().
			PaddingLeft(2)

	selectedItemStyle = lipgloss.NewStyle().
				PaddingLeft(1).
				Foreground(colorPrimary).
				Bold(true).
				SetString("▸ ")

	activeProfileStyle = lipgloss.NewStyle().
				Foreground(colorSuccess)

	// Badge styles
	badgeSSO = lipgloss.NewStyle().
			Foreground(colorInfo).
			Bold(true).
			SetString("SSO")

	badgeRole = lipgloss.NewStyle().
			Foreground(colorWarning).
			Bold(true).
			SetString("ROLE")

	badgeStatic = lipgloss.NewStyle().
			Foreground(colorSuccess).
			Bold(true).
			SetString("STATIC")

	badgeUnknown = lipgloss.NewStyle().
			Foreground(colorSecondary).
			SetString("UNKNOWN")

	// Detail panel
	detailLabelStyle = lipgloss.NewStyle().
				Foreground(colorSecondary).
				Width(16)

	detailValueStyle = lipgloss.NewStyle().
				Foreground(colorWhite)

	detailTitleStyle = lipgloss.NewStyle().
				Foreground(colorPrimary).
				Bold(true).
				MarginBottom(1)

	// Status bar
	statusBarStyle = lipgloss.NewStyle().
			Foreground(colorSecondary).
			Padding(0, 1).
			MarginTop(1)

	// Search
	searchStyle = lipgloss.NewStyle().
			Foreground(colorPrimary).
			Bold(true)

	// Title
	titleStyle = lipgloss.NewStyle().
			Foreground(colorPrimary).
			Bold(true).
			MarginBottom(1)

	// Result message shown after selecting a profile
	resultStyle = lipgloss.NewStyle().
			Foreground(colorSuccess).
			Bold(true)

	// Inline badge styles for list panel
	badgeInlineSSO = lipgloss.NewStyle().
			Foreground(colorInfo).
			SetString(" [SSO]")

	badgeInlineRole = lipgloss.NewStyle().
			Foreground(colorWarning).
			SetString(" [ROLE]")

	badgeInlineStatic = lipgloss.NewStyle().
			Foreground(colorSuccess).
			SetString(" [STATIC]")

	badgeInlineUnknown = lipgloss.NewStyle().
				Foreground(colorSecondary).
				SetString(" [?]")

	// Divider between active profile and the rest
	dividerStyle = lipgloss.NewStyle().
			Foreground(colorMuted)
)
