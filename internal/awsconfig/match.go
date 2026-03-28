package awsconfig

import (
	"fmt"
	"sort"
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
// 3. Case-insensitive prefix match
// 4. Fuzzy match (all query chars appear in order)
// Returns an error with suggestions if no match or ambiguous.
func MatchProfile(name string, profiles []models.Profile) (*MatchResult, error) {
	query := strings.TrimSpace(name)
	if query == "" {
		return nil, fmt.Errorf("profile name is required")
	}

	lower := strings.ToLower(query)

	// 1. Exact match.
	for i := range profiles {
		if profiles[i].Name == query {
			return &MatchResult{Profile: &profiles[i]}, nil
		}
	}

	// 2. Case-insensitive match.
	var ciMatches []int
	for i := range profiles {
		if strings.ToLower(profiles[i].Name) == lower {
			ciMatches = append(ciMatches, i)
		}
	}
	if len(ciMatches) == 1 {
		return &MatchResult{Profile: &profiles[ciMatches[0]]}, nil
	}
	if len(ciMatches) > 1 {
		return nil, ambiguousProfileError(query, profileNames(profiles, ciMatches))
	}

	// 3. Case-insensitive prefix match.
	var prefixMatches []int
	for i := range profiles {
		if strings.HasPrefix(strings.ToLower(profiles[i].Name), lower) {
			prefixMatches = append(prefixMatches, i)
		}
	}

	if len(prefixMatches) == 1 {
		return &MatchResult{Profile: &profiles[prefixMatches[0]]}, nil
	}
	if len(prefixMatches) > 1 {
		return nil, ambiguousProfileError(query, rankedNames(query, profiles, prefixMatches))
	}

	// 4. Fuzzy match.
	var fuzzyMatches []int
	for i := range profiles {
		if fuzzyContains(strings.ToLower(profiles[i].Name), lower) {
			fuzzyMatches = append(fuzzyMatches, i)
		}
	}

	if len(fuzzyMatches) == 1 {
		return &MatchResult{Profile: &profiles[fuzzyMatches[0]]}, nil
	}
	if len(fuzzyMatches) > 1 {
		return nil, ambiguousProfileError(query, rankedNames(query, profiles, fuzzyMatches))
	}

	suggestions := rankedNames(query, profiles, nearestProfileIndexes(lower, profiles))
	if len(suggestions) > 0 {
		return nil, fmt.Errorf("profile %q not found. nearest matches: %s", query, strings.Join(suggestions, ", "))
	}

	return nil, fmt.Errorf("profile %q not found", query)
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

func ambiguousProfileError(query string, suggestions []string) error {
	return fmt.Errorf("profile %q is ambiguous. matches: %s", query, strings.Join(suggestions, ", "))
}

func profileNames(profiles []models.Profile, indexes []int) []string {
	names := make([]string, 0, len(indexes))
	for _, idx := range indexes {
		names = append(names, profiles[idx].Name)
	}
	return names
}

func rankedNames(query string, profiles []models.Profile, indexes []int) []string {
	if len(indexes) == 0 {
		return nil
	}

	type candidate struct {
		name  string
		score int
	}

	queryLower := strings.ToLower(query)
	candidates := make([]candidate, 0, len(indexes))
	for _, idx := range indexes {
		name := profiles[idx].Name
		candidates = append(candidates, candidate{
			name:  name,
			score: matchScore(queryLower, strings.ToLower(name)),
		})
	}

	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].score == candidates[j].score {
			return candidates[i].name < candidates[j].name
		}
		return candidates[i].score < candidates[j].score
	})

	limit := len(candidates)
	if limit > 5 {
		limit = 5
	}

	names := make([]string, 0, limit)
	for i := 0; i < limit; i++ {
		names = append(names, candidates[i].name)
	}
	return names
}

func nearestProfileIndexes(query string, profiles []models.Profile) []int {
	indexes := make([]int, 0, len(profiles))
	for i := range profiles {
		name := strings.ToLower(profiles[i].Name)
		if strings.Contains(name, query) || fuzzyContains(name, query) {
			indexes = append(indexes, i)
		}
	}
	return indexes
}

func matchScore(query, candidate string) int {
	if candidate == query {
		return 0
	}
	if strings.HasPrefix(candidate, query) {
		return len(candidate) - len(query)
	}
	if idx := strings.Index(candidate, query); idx >= 0 {
		return 100 + idx + (len(candidate) - len(query))
	}
	return 1000 + len(candidate)
}
