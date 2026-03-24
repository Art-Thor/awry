package session

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Art-Thor/awry/pkg/models"
)

const expiringSoonWindow = 15 * time.Minute

// Status represents the current state of a profile session.
type Status string

const (
	StatusNotApplicable Status = "not_applicable"
	StatusUnknown       Status = "unknown"
	StatusActive        Status = "active"
	StatusExpiringSoon  Status = "expiring_soon"
	StatusExpired       Status = "expired"
)

// Source describes where session lifetime information came from.
type Source string

const (
	SourceNone Source = "none"
	SourceSSO  Source = "sso"
)

// Info contains normalized session lifetime details for a profile.
type Info struct {
	Status    Status
	Source    Source
	ExpiresAt time.Time
	Remaining time.Duration
}

type ssoTokenCache struct {
	StartURL  string `json:"startUrl"`
	IssuerURL string `json:"issuerUrl"`
	Region    string `json:"region"`
	ExpiresAt string `json:"expiresAt"`
}

// DefaultSSOCachePath returns the default AWS SSO token cache path.
func DefaultSSOCachePath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".aws", "sso", "cache")
}

// ResolveProfile returns session information for a profile using local AWS caches.
func ResolveProfile(profile models.Profile) (Info, error) {
	return ResolveProfileFrom(profile, DefaultSSOCachePath(), time.Now())
}

// ResolveProfileFrom returns session information for a profile using the given cache path and time.
func ResolveProfileFrom(profile models.Profile, cacheDir string, now time.Time) (Info, error) {
	if profile.Type != models.ProfileTypeSSO || profile.SSOStartURL == "" {
		return Info{Status: StatusNotApplicable, Source: SourceNone}, nil
	}

	token, err := loadMatchingSSOToken(cacheDir, profile)
	if err != nil {
		if errors.Is(err, errNoMatchingSSOToken) || errors.Is(err, os.ErrNotExist) {
			return Info{Status: StatusUnknown, Source: SourceSSO}, nil
		}
		return Info{}, err
	}

	expiresAt, err := time.Parse(time.RFC3339, token.ExpiresAt)
	if err != nil {
		return Info{}, fmt.Errorf("parsing SSO session expiry for %q: %w", profile.Name, err)
	}

	remaining := expiresAt.Sub(now)
	status := sessionStatus(remaining)
	if status == StatusExpired {
		remaining = 0
	}

	return Info{
		Status:    status,
		Source:    SourceSSO,
		ExpiresAt: expiresAt,
		Remaining: remaining,
	}, nil
}

var errNoMatchingSSOToken = errors.New("no matching SSO token")

func loadMatchingSSOToken(cacheDir string, profile models.Profile) (ssoTokenCache, error) {
	entries, err := os.ReadDir(cacheDir)
	if err != nil {
		return ssoTokenCache{}, err
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		path := filepath.Join(cacheDir, entry.Name())
		content, err := os.ReadFile(path)
		if err != nil {
			return ssoTokenCache{}, fmt.Errorf("reading SSO cache %s: %w", path, err)
		}

		var token ssoTokenCache
		if err := json.Unmarshal(content, &token); err != nil {
			continue
		}

		if matchesProfile(token, profile) {
			return token, nil
		}
	}

	return ssoTokenCache{}, errNoMatchingSSOToken
}

func matchesProfile(token ssoTokenCache, profile models.Profile) bool {
	profileURL := normalizeURL(profile.SSOStartURL)
	return token.Region == profile.SSORegion && (normalizeURL(token.StartURL) == profileURL || normalizeURL(token.IssuerURL) == profileURL)
}

func normalizeURL(value string) string {
	value = strings.TrimSpace(value)
	return strings.TrimRight(value, "/")
}

func sessionStatus(remaining time.Duration) Status {
	if remaining <= 0 {
		return StatusExpired
	}
	if remaining <= expiringSoonWindow {
		return StatusExpiringSoon
	}
	return StatusActive
}
