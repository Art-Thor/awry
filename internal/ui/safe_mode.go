package ui

import (
	"strings"

	"github.com/Art-Thor/awry/internal/config"
	"github.com/Art-Thor/awry/pkg/models"
)

func (m Model) safeModeMatch(profile models.Profile) (config.SafeModeMatch, bool) {
	return config.MatchProductionProfile(profile.Name, m.productionPatterns)
}

func (m Model) requiresProductionConfirmation(profile models.Profile) bool {
	if !m.confirmProduction || profile.Name == m.currentProfile {
		return false
	}

	_, ok := m.safeModeMatch(profile)
	return ok
}

func (m *Model) confirmSelection(profile models.Profile) {
	m.confirmingSafe = m.requiresProductionConfirmation(profile)
}

func (m *Model) selectProfile(profile models.Profile) {
	_ = m.recordRecent(profile.Name)
	m.selected = &profile
	m.confirmingSafe = false
}

func (m Model) safeModeListBadge(profile models.Profile) string {
	match, ok := m.safeModeMatch(profile)
	if !ok {
		return ""
	}

	if match.Level == config.DangerLevelDanger {
		return safeModeBadgeDanger.String()
	}

	return safeModeBadgeWarning.String()
}

func (m Model) safeModeBanner(profile models.Profile) string {
	match, ok := m.safeModeMatch(profile)
	if !ok {
		return ""
	}

	message := "Safe Mode: production-like profile"
	if m.confirmingSafe && m.requiresProductionConfirmation(profile) {
		message = "Safe Mode: press Enter again to confirm, or Esc to cancel"
	} else if m.requiresProductionConfirmation(profile) {
		message = "Safe Mode: Enter will ask for confirmation before switching"
	}

	details := "Matched pattern: " + match.Pattern
	if match.Level == config.DangerLevelDanger {
		return safeModeBannerDangerStyle.Render(strings.ToUpper(message)) + "\n" + safeModeBannerNoteStyle.Render(details)
	}

	return safeModeBannerWarningStyle.Render(strings.ToUpper(message)) + "\n" + safeModeBannerNoteStyle.Render(details)
}

func (m Model) safeModeValue(profile models.Profile) string {
	match, ok := m.safeModeMatch(profile)
	if !ok {
		return "Off"
	}

	level := "Warning"
	if match.Level == config.DangerLevelDanger {
		level = "Danger"
	}

	if m.requiresProductionConfirmation(profile) {
		return level + " - confirmation required"
	}

	return level + " - production profile"
}
