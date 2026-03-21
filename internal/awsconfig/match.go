package awsconfig

import (
	"fmt"
	"strings"

	"github.com/Art-Thor/awry/pkg/models"
)

// MatchResult represents the outcome of a profile matching attempt.
type MatchResult struct {
	Profile     *models.Profile
	Suggestions []string
}

// MatchProfile finds a profile by name using a cascading strategy:
// 1. Exact match
// 2. Case-insensitive match
// 3. Fuzzy match (all query chars appear in order)
// Returns an error with suggestions if no match or ambiguous.
func MatchProfile(name string, profiles []models.Profile) (*MatchResult, error) {
	if name == "" {
		return nil, fmt.Errorf("profile name is required")
	}

	// 1. Exact match.
	for i := range profiles {
		if profiles[i].Name == name {
			return &MatchResult{Profile: &profiles[i]}, nil
		}
	}

	// 2. Case-insensitive match.
	var ciMatches []int
	lower := strings.ToLower(name)
	for i := range profiles {
		if strings.ToLower(profiles[i].Name) == lower {
			ciMatches = append(ciMatches, i)
		}
	}
	if len(ciMatches) == 1 {
		return &MatchResult{Profile: &profiles[ciMatches[0]]}, nil
	}

	// 3. Fuzzy match.
	var fuzzyMatches []int
	for i := range profiles {
		if fuzzyContains(strings.ToLower(profiles[i].Name), lower) {
			fuzzyMatches = append(fuzzyMatches, i)
		}
	}

	if len(fuzzyMatches) == 1 {
		return &MatchResult{Profile: &profiles[fuzzyMatches[0]]}, nil
	}

	// Ambiguous or not found — collect suggestions.
	candidates := fuzzyMatches
	if len(candidates) == 0 && len(ciMatches) > 0 {
		candidates = ciMatches
	}

	suggestions := make([]string, len(candidates))
	for i, idx := range candidates {
		suggestions[i] = profiles[idx].Name
	}

	if len(suggestions) > 0 {
		return nil, fmt.Errorf("ambiguous profile %q, did you mean one of: %s", name, strings.Join(suggestions, ", "))
	}

	return nil, fmt.Errorf("profile %q not found", name)
}

// fuzzyContains checks if all characters in query appear in order within s.
func fuzzyContains(s, query string) bool {
	qi := 0
	for i := 0; i < len(s) && qi < len(query); i++ {
		if s[i] == query[qi] {
			qi++
		}
	}
	return qi == len(query)
}
