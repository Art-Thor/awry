package ui

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/Art-Thor/awry/pkg/models"
)

func TestRecordRecentPersistsAndReorders(t *testing.T) {
	m := Model{
		profiles: []models.Profile{
			{Name: "dev"},
			{Name: "prod"},
			{Name: "sandbox"},
			{Name: "staging"},
		},
		filtered:       []models.Profile{{Name: "dev"}, {Name: "prod"}, {Name: "sandbox"}, {Name: "staging"}},
		currentProfile: "dev",
		favorites:      map[string]struct{}{"prod": {}},
		configPath:     filepath.Join(t.TempDir(), "config.yaml"),
	}

	if err := m.recordRecent("staging"); err != nil {
		t.Fatalf("recordRecent() unexpected error: %v", err)
	}

	if !m.isRecent("staging") {
		t.Fatal("expected staging to be recent")
	}

	if got := m.profiles[2].Name; got != "staging" {
		t.Fatalf("expected recent after favorites, got %q", got)
	}
}

func TestRenderDetailShowsRecentField(t *testing.T) {
	m := Model{
		profiles: []models.Profile{{Name: "sandbox", Type: models.ProfileTypeStatic, HasCredentials: true}},
		filtered: []models.Profile{{Name: "sandbox", Type: models.ProfileTypeStatic, HasCredentials: true}},
		recents:  []string{"sandbox"},
	}

	view := m.renderDetail(80)
	if !strings.Contains(view, "Recent") || !strings.Contains(view, "Yes") {
		t.Fatalf("expected recent field in detail view\n%s", view)
	}
}
