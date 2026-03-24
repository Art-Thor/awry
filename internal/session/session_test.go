package session

import (
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/Art-Thor/awry/pkg/models"
)

func TestResolveProfileFrom(t *testing.T) {
	now := time.Date(2026, time.March, 23, 13, 0, 0, 0, time.UTC)
	profile := models.Profile{
		Name:        "sandbox-admin",
		Type:        models.ProfileTypeSSO,
		SSOStartURL: "https://example.awsapps.com/start",
		SSORegion:   "us-east-1",
	}

	tests := []struct {
		name       string
		cacheDir   string
		profile    models.Profile
		wantStatus Status
		wantSource Source
		wantExpiry time.Time
		wantRemain time.Duration
		wantErr    bool
	}{
		{
			name:       "active session",
			cacheDir:   filepath.Join("testdata", "active"),
			profile:    profile,
			wantStatus: StatusActive,
			wantSource: SourceSSO,
			wantExpiry: time.Date(2026, time.March, 23, 15, 30, 0, 0, time.UTC),
			wantRemain: 2*time.Hour + 30*time.Minute,
		},
		{
			name:       "expiring soon session",
			cacheDir:   filepath.Join("testdata", "expiring"),
			profile:    profile,
			wantStatus: StatusExpiringSoon,
			wantSource: SourceSSO,
			wantExpiry: time.Date(2026, time.March, 23, 13, 10, 0, 0, time.UTC),
			wantRemain: 10 * time.Minute,
		},
		{
			name:       "expired session",
			cacheDir:   filepath.Join("testdata", "expired"),
			profile:    profile,
			wantStatus: StatusExpired,
			wantSource: SourceSSO,
			wantExpiry: time.Date(2026, time.March, 23, 12, 0, 0, 0, time.UTC),
			wantRemain: 0,
		},
		{
			name:       "missing cache returns unknown",
			cacheDir:   filepath.Join("testdata", "missing"),
			profile:    profile,
			wantStatus: StatusUnknown,
			wantSource: SourceSSO,
		},
		{
			name:       "unmatched cache returns unknown",
			cacheDir:   filepath.Join("testdata", "unmatched"),
			profile:    profile,
			wantStatus: StatusUnknown,
			wantSource: SourceSSO,
		},
		{
			name:     "malformed matching cache returns error",
			cacheDir: filepath.Join("testdata", "malformed"),
			profile:  profile,
			wantErr:  true,
		},
		{
			name:       "non sso profile is not applicable",
			cacheDir:   filepath.Join("testdata", "active"),
			profile:    models.Profile{Name: "static", Type: models.ProfileTypeStatic},
			wantStatus: StatusNotApplicable,
			wantSource: SourceNone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := ResolveProfileFrom(tt.profile, tt.cacheDir, now)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if info.Status != tt.wantStatus {
				t.Fatalf("status = %q, want %q", info.Status, tt.wantStatus)
			}
			if info.Source != tt.wantSource {
				t.Fatalf("source = %q, want %q", info.Source, tt.wantSource)
			}
			if !info.ExpiresAt.Equal(tt.wantExpiry) {
				t.Fatalf("expiresAt = %v, want %v", info.ExpiresAt, tt.wantExpiry)
			}
			if info.Remaining != tt.wantRemain {
				t.Fatalf("remaining = %v, want %v", info.Remaining, tt.wantRemain)
			}
		})
	}
}

func TestLoadMatchingSSOToken(t *testing.T) {
	profile := models.Profile{
		Name:        "sandbox-admin",
		Type:        models.ProfileTypeSSO,
		SSOStartURL: "https://example.awsapps.com/start/",
		SSORegion:   "us-east-1",
	}

	t.Run("matches issuer url and trims trailing slash", func(t *testing.T) {
		token, err := loadMatchingSSOToken(filepath.Join("testdata", "issuer"), profile)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if token.IssuerURL != "https://example.awsapps.com/start" {
			t.Fatalf("unexpected issuer url: %q", token.IssuerURL)
		}
	})

	t.Run("returns no match sentinel", func(t *testing.T) {
		_, err := loadMatchingSSOToken(filepath.Join("testdata", "unmatched"), profile)
		if !errors.Is(err, errNoMatchingSSOToken) {
			t.Fatalf("expected no matching token error, got %v", err)
		}
	})
}
