package ui

import (
	"strings"
	"testing"

	"github.com/Art-Thor/awry/pkg/models"
)

func TestIsRiskyProfileDefaultPatterns(t *testing.T) {
	m := Model{}

	for _, name := range []string{"prod-admin", "production-readonly", "live-eu"} {
		if !m.isRiskyProfile(name) {
			t.Fatalf("expected %q to be risky", name)
		}
	}

	if m.isRiskyProfile("sandbox") {
		t.Fatal("did not expect sandbox to be risky")
	}
}

func TestIsRiskyProfileCustomPatterns(t *testing.T) {
	m := Model{configRiskPatterns: []string{"critical", "payments"}}

	if !m.isRiskyProfile("payments-admin") {
		t.Fatal("expected custom risk pattern to match")
	}
	if m.isRiskyProfile("prod-admin") {
		t.Fatal("did not expect default pattern match when custom patterns are configured")
	}
}

func TestRenderDetailShowsRiskField(t *testing.T) {
	m := Model{
		profiles: []models.Profile{{Name: "prod-admin", Type: models.ProfileTypeStatic, HasCredentials: true}},
		filtered: []models.Profile{{Name: "prod-admin", Type: models.ProfileTypeStatic, HasCredentials: true}},
	}

	view := m.renderDetail(80)
	if !strings.Contains(view, "Risk") || !strings.Contains(view, "Production-like") {
		t.Fatalf("expected risk field in detail view\n%s", view)
	}
}
