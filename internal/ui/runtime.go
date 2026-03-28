package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Art-Thor/awry/internal/session"
	"github.com/Art-Thor/awry/pkg/models"
)

func (m Model) refreshActiveRuntimeCmd() tea.Cmd {
	active, ok := m.activeProfile()
	if !ok {
		return nil
	}

	return tea.Batch(loadSessionCmd(active), loadIdentityCmd(active.Name))
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
