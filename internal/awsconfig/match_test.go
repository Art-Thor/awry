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
	result, err := MatchProfile("stag", profiles)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Profile.Name != "staging" {
		t.Errorf("expected staging, got %s", result.Profile.Name)
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
}

func TestMatchProfile_NotFound(t *testing.T) {
	profiles := testProfiles()
	_, err := MatchProfile("nonexistent", profiles)
	if err == nil {
		t.Fatal("expected error for not found")
	}
}

func TestMatchProfile_Empty(t *testing.T) {
	_, err := MatchProfile("", testProfiles())
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}
