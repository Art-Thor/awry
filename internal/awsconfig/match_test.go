package awsconfig

import (
	"testing"

	"github.com/Art-Thor/awry/pkg/models"
)

func testProfiles() []models.Profile {
	return []models.Profile{
		{Name: "dev-account"},
		{Name: "Dev-Staging"},
		{Name: "production"},
		{Name: "prod-readonly"},
		{Name: "staging"},
	}
}

func TestMatchProfile_Exact(t *testing.T) {
	profiles := testProfiles()
	result, err := MatchProfile("production", profiles)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Profile.Name != "production" {
		t.Errorf("expected production, got %s", result.Profile.Name)
	}
}

func TestMatchProfile_CaseInsensitive(t *testing.T) {
	profiles := testProfiles()
	result, err := MatchProfile("PRODUCTION", profiles)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Profile.Name != "production" {
		t.Errorf("expected production, got %s", result.Profile.Name)
	}
}

func TestMatchProfile_Fuzzy(t *testing.T) {
	profiles := testProfiles()
	result, err := MatchProfile("devac", profiles)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Profile.Name != "dev-account" {
		t.Errorf("expected dev-account, got %s", result.Profile.Name)
	}
}

func TestMatchProfile_PrefixPreferredOverFuzzy(t *testing.T) {
	profiles := testProfiles()
	result, err := MatchProfile("produc", profiles)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Profile.Name != "production" {
		t.Errorf("expected production, got %s", result.Profile.Name)
	}
}

func TestMatchProfile_Ambiguous(t *testing.T) {
	profiles := append(testProfiles(), models.Profile{Name: "prod-admin"})
	_, err := MatchProfile("prod", profiles)
	if err == nil {
		t.Fatal("expected error for ambiguous match")
	}
	if got := err.Error(); got != `profile "prod" is ambiguous. matches: prod-admin, production, prod-readonly` {
		t.Fatalf("unexpected error: %s", got)
	}
}

func TestMatchProfile_NotFound(t *testing.T) {
	profiles := testProfiles()
	_, err := MatchProfile("nonexistent", profiles)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if got := err.Error(); got != `profile "nonexistent" not found` {
		t.Fatalf("unexpected error: %s", got)
	}
}

func TestMatchProfile_AmbiguousFuzzySuggestions(t *testing.T) {
	profiles := testProfiles()
	_, err := MatchProfile("deva", profiles)
	if err == nil {
		t.Fatal("expected error for ambiguous match")
	}
	if got := err.Error(); got != `profile "deva" is ambiguous. matches: Dev-Staging, dev-account` {
		t.Fatalf("unexpected error: %s", got)
	}
}

func TestMatchProfile_Empty(t *testing.T) {
	_, err := MatchProfile("", testProfiles())
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestMatchProfile_WhitespaceOnly(t *testing.T) {
	_, err := MatchProfile("   ", testProfiles())
	if err == nil {
		t.Fatal("expected error for whitespace-only name")
	}
	if got := err.Error(); got != `profile name is required` {
		t.Fatalf("unexpected error: %s", got)
	}
}
