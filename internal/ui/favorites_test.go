package ui

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/Art-Thor/awry/pkg/models"
)

func TestToggleFavoritePersistsAndReorders(t *testing.T) {
	m := Model{
		profiles: []models.Profile{
			{Name: "dev"},
			{Name: "prod"},
			{Name: "sandbox"},
		},
		filtered:       []models.Profile{{Name: "dev"}, {Name: "prod"}, {Name: "sandbox"}},
		currentProfile: "dev",
		favorites:      map[string]struct{}{},
		configPath:     filepath.Join(t.TempDir(), "config.yaml"),
	}

	if err := m.toggleFavorite("sandbox"); err != nil {
		t.Fatalf("toggleFavorite() unexpected error: %v", err)
	}

	if !m.isFavorite("sandbox") {
		t.Fatal("expected sandbox to be favorite")
	}

	if got := m.profiles[1].Name; got != "sandbox" {
		t.Fatalf("expected favorite after current profile, got %q", got)
	}

	if err := m.toggleFavorite("sandbox"); err != nil {
		t.Fatalf("toggleFavorite() unexpected error: %v", err)
	}

	if m.isFavorite("sandbox") {
		t.Fatal("expected sandbox favorite to be removed")
	}
}

func TestRenderDetailShowsFavoriteField(t *testing.T) {
	m := Model{
		profiles:  []models.Profile{{Name: "sandbox", Type: models.ProfileTypeStatic, HasCredentials: true}},
		filtered:  []models.Profile{{Name: "sandbox", Type: models.ProfileTypeStatic, HasCredentials: true}},
		favorites: map[string]struct{}{"sandbox": {}},
	}

	view := m.renderDetail(80)
	if !strings.Contains(view, "Favorite") || !strings.Contains(view, "Yes") {
		t.Fatalf("expected favorite field in detail view\n%s", view)
	}
}
