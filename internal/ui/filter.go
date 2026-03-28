package ui

import (
	"strings"

	"github.com/Art-Thor/awry/pkg/models"
)

func (m *Model) applyFilter() {
	if m.searchQuery == "" {
		m.filtered = m.profiles
		m.cursor = 0
		return
	}

	query := strings.ToLower(m.searchQuery)
	var result []models.Profile
	for _, p := range m.profiles {
		if fuzzyMatch(strings.ToLower(p.Name), query) {
			result = append(result, p)
		}
	}
	m.filtered = result
	m.cursor = 0
}

// pinActiveToTop moves the active profile to the first position in the list.
func (m *Model) pinActiveToTop() {
	if m.currentProfile == "" {
		return
	}
	for i, p := range m.profiles {
		if p.Name == m.currentProfile && i > 0 {
			m.profiles = append([]models.Profile{p}, append(m.profiles[:i], m.profiles[i+1:]...)...)
			return
		}
	}
}

func (m Model) activeProfile() (models.Profile, bool) {
	if m.currentProfile == "" {
		return models.Profile{}, false
	}
	for _, profile := range m.profiles {
		if profile.Name == m.currentProfile {
			return profile, true
		}
	}
	return models.Profile{}, false
}

// fuzzyMatch checks if all characters in query appear in order within s.
func fuzzyMatch(s, query string) bool {
	qi := 0
	for i := 0; i < len(s) && qi < len(query); i++ {
		if s[i] == query[qi] {
			qi++
		}
	}
	return qi == len(query)
}
