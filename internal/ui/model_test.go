package ui

import (
	"errors"
	"strings"
	"testing"
	"time"

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
		{name: "error", err: errors.New("boom"), want: "boom"},
		{name: "no expiry", info: &session.Info{Status: session.StatusNotApplicable}, want: "No expiry"},
		{name: "unknown", info: &session.Info{Status: session.StatusUnknown}, want: "Unknown"},
		{name: "expired", info: &session.Info{Status: session.StatusExpired}, want: "Expired"},
		{name: "expiring", info: &session.Info{Status: session.StatusExpiringSoon, Remaining: 10 * time.Minute}, want: "10m left (expiring soon)"},
		{name: "active", info: &session.Info{Status: session.StatusActive, Remaining: 95 * time.Minute}, want: "1h 35m left"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := Model{sessionInfo: tt.info, sessionErr: tt.err}
			if got := m.sessionStatusValue(); got != tt.want {
				t.Fatalf("sessionStatusValue() = %q, want %q", got, tt.want)
			}
		})
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
	if !strings.Contains(view, "AWS SSO session expired") {
		t.Fatalf("expected identity error in detail view\n%s", view)
	}
}
