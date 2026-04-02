package ui

import (
	"strings"
	"testing"

	"github.com/Art-Thor/awry/pkg/models"
)

func TestRenderListShowsSectionHeaders(t *testing.T) {
	m := Model{
		profiles: []models.Profile{
			{Name: "dev"},
			{Name: "prod"},
			{Name: "staging"},
			{Name: "sandbox"},
		},
		filtered: []models.Profile{
			{Name: "dev"},
			{Name: "prod"},
			{Name: "staging"},
			{Name: "sandbox"},
		},
		currentProfile: "dev",
		favorites:      map[string]struct{}{"prod": {}},
		recents:        []string{"staging"},
		height:         20,
	}

	view := m.renderList(40)
	for _, section := range []string{"Current", "Favorites", "Recent", "All Profiles"} {
		if !strings.Contains(view, section) {
			t.Fatalf("expected section %q in list\n%s", section, view)
		}
	}
	if strings.Count(view, "Favorites") != 1 {
		t.Fatalf("expected one Favorites header\n%s", view)
	}
}

func TestSectionHeadersHiddenWhileSearching(t *testing.T) {
	m := Model{
		filtered:    []models.Profile{{Name: "prod"}},
		favorites:   map[string]struct{}{"prod": {}},
		searching:   true,
		searchQuery: "pr",
		height:      20,
	}

	view := m.renderList(40)
	if strings.Contains(view, "Favorites") || strings.Contains(view, "All Profiles") {
		t.Fatalf("expected no section headers while searching\n%s", view)
	}
}
