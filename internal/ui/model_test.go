package ui

import (
	"errors"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Art-Thor/awry/internal/identity"
	"github.com/Art-Thor/awry/internal/session"
	"github.com/Art-Thor/awry/pkg/models"
)

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name string
		in   time.Duration
		want string
	}{
		{name: "zero", in: 0, want: "0m"},
		{name: "minutes", in: 47 * time.Minute, want: "47m"},
		{name: "hours", in: 2 * time.Hour, want: "2h"},
		{name: "hours and minutes", in: 2*time.Hour + 15*time.Minute, want: "2h 15m"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatDuration(tt.in); got != tt.want {
				t.Fatalf("formatDuration(%v) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestSessionStatusValue(t *testing.T) {
	tests := []struct {
		name string
		info *session.Info
		err  error
		want string
	}{
		{name: "loading", want: "Loading..."},
		{name: "error", err: errors.New("boom"), want: "Unavailable"},
		{name: "no expiry", info: &session.Info{Status: session.StatusNotApplicable}, want: "No expiry"},
		{name: "unknown", info: &session.Info{Status: session.StatusUnknown}, want: "Unknown - run aws sso login if needed"},
		{name: "expired", info: &session.Info{Status: session.StatusExpired}, want: "Expired - refresh credentials"},
		{name: "expiring", info: &session.Info{Status: session.StatusExpiringSoon, Remaining: 10 * time.Minute}, want: "10m left (expiring soon)"},
		{name: "active", info: &session.Info{Status: session.StatusActive, Remaining: 95 * time.Minute}, want: "1h 35m left"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := Model{sessionInfo: tt.info, sessionErr: tt.err}
			if got := m.sessionStatusValue(models.Profile{Type: models.ProfileTypeSSO}); got != tt.want {
				t.Fatalf("sessionStatusValue() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestIdentityStatusValue(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{name: "expired", err: errors.New("AWS SSO session expired for profile \"sandbox\""), want: "Expired - run aws sso login"},
		{name: "no creds", err: errors.New("no valid AWS credentials available for profile \"sandbox\""), want: "No credentials available"},
		{name: "other", err: errors.New("boom"), want: "Unavailable"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := identityStatusValue(tt.err); got != tt.want {
				t.Fatalf("identityStatusValue() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestActiveRuntimeBadge(t *testing.T) {
	tests := []struct {
		name string
		m    Model
		want string
	}{
		{name: "loading", m: Model{}, want: " [LOAD]"},
		{name: "error", m: Model{sessionErr: errors.New("boom")}, want: " [CHECK]"},
		{name: "no creds", m: Model{identityErr: errors.New("no valid AWS credentials available for profile \"sandbox\"")}, want: " [NO CREDS]"},
		{name: "active", m: Model{sessionInfo: &session.Info{Status: session.StatusActive}}, want: " [READY]"},
		{name: "soon", m: Model{sessionInfo: &session.Info{Status: session.StatusExpiringSoon}}, want: " [EXPIRING]"},
		{name: "expired", m: Model{sessionInfo: &session.Info{Status: session.StatusExpired}}, want: " [EXPIRED]"},
		{name: "unknown", m: Model{sessionInfo: &session.Info{Status: session.StatusUnknown}}, want: " [UNKNOWN]"},
		{name: "not applicable", m: Model{sessionInfo: &session.Info{Status: session.StatusNotApplicable}}, want: " [READY]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.m.activeRuntimeBadge(); got != tt.want {
				t.Fatalf("activeRuntimeBadge() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRenderDetailShowsShellSafeExportPreview(t *testing.T) {
	m := Model{
		profiles: []models.Profile{{Name: "team's sandbox", Type: models.ProfileTypeStatic, HasCredentials: true}},
		filtered: []models.Profile{{Name: "team's sandbox", Type: models.ProfileTypeStatic, HasCredentials: true}},
	}

	view := m.renderDetail(80)
	if !strings.Contains(view, `export AWS_PROFILE='team'"'"'s sandbox'`) {
		t.Fatalf("expected shell-safe export preview in detail view\n%s", view)
	}
}

func TestRenderDetailForActiveProfile(t *testing.T) {
	m := Model{
		profiles:       []models.Profile{{Name: "sandbox-admin", Type: models.ProfileTypeSSO, SSOStartURL: "https://example.awsapps.com/start", SSORegion: "us-east-1"}},
		filtered:       []models.Profile{{Name: "sandbox-admin", Type: models.ProfileTypeSSO, SSOStartURL: "https://example.awsapps.com/start", SSORegion: "us-east-1"}},
		currentProfile: "sandbox-admin",
		sessionInfo:    &session.Info{Status: session.StatusActive, Remaining: 47 * time.Minute},
		identity:       &identity.Identity{Profile: "sandbox-admin", AccountID: "123456789012", ARN: "arn:aws:sts::123456789012:assumed-role/Admin/sandbox-admin", Principal: "sandbox-admin"},
	}

	view := m.renderDetail(80)

	checks := []string{
		"Session",
		"47m left",
		"Account ID",
		"123456789012",
		"Principal",
		"sandbox-admin",
		"Currently active",
	}

	for _, check := range checks {
		if !strings.Contains(view, check) {
			t.Fatalf("expected detail view to contain %q\n%s", check, view)
		}
	}
}

func TestRenderDetailShowsIdentityError(t *testing.T) {
	m := Model{
		profiles:       []models.Profile{{Name: "sandbox-admin", Type: models.ProfileTypeSSO}},
		filtered:       []models.Profile{{Name: "sandbox-admin", Type: models.ProfileTypeSSO}},
		currentProfile: "sandbox-admin",
		identityErr:    errors.New("AWS SSO session expired"),
	}

	view := m.renderDetail(80)
	if !strings.Contains(view, "Expired - run aws sso login") {
		t.Fatalf("expected identity error in detail view\n%s", view)
	}
}

func TestRenderStatusBarIncludesRefresh(t *testing.T) {
	view := Model{}.renderStatusBar()
	if !strings.Contains(view, "r refresh") {
		t.Fatalf("expected refresh hint in status bar\n%s", view)
	}
	if !strings.Contains(view, "? help") {
		t.Fatalf("expected help hint in status bar\n%s", view)
	}
}

func TestHelpOverlayToggle(t *testing.T) {
	m := Model{}
	updated, _ := m.handleKey(key("?"))
	opened := updated.(Model)
	if !opened.helpVisible {
		t.Fatal("expected help overlay to open")
	}

	updated, _ = opened.handleKey(key("esc"))
	closed := updated.(Model)
	if closed.helpVisible {
		t.Fatal("expected help overlay to close")
	}
}

func TestRenderHelpOverlay(t *testing.T) {
	view := Model{}.renderHelpOverlay()
	checks := []string{"Keyboard Help", "r", "Refresh active session and identity", "Press ? or Esc to close"}
	for _, check := range checks {
		if !strings.Contains(view, check) {
			t.Fatalf("expected help overlay to contain %q\n%s", check, view)
		}
	}
}

func key(value string) tea.KeyMsg {
	if value == "esc" {
		return tea.KeyMsg{Type: tea.KeyEsc}
	}
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(value)}
}
