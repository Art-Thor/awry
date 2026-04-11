package config

import (
	"regexp"
	"strings"
)

// DangerLevel represents the severity of a production profile match.
type DangerLevel string

const (
	DangerLevelWarning DangerLevel = "warning"
	DangerLevelDanger  DangerLevel = "danger"
)

// SafeModeMatch describes which pattern marked a profile as production-like.
type SafeModeMatch struct {
	Pattern string
	Level   DangerLevel
}

var defaultProductionPatterns = []string{"prod", "production", "live"}

// DefaultProductionPatterns returns the built-in production profile patterns.
func DefaultProductionPatterns() []string {
	return append([]string(nil), defaultProductionPatterns...)
}

// MatchProductionProfile returns the first configured production-like match.
func MatchProductionProfile(name string, patterns []string) (SafeModeMatch, bool) {
	lowerName := strings.ToLower(name)

	for _, pattern := range effectiveProductionPatterns(patterns) {
		normalized := strings.ToLower(strings.TrimSpace(pattern))
		if normalized == "" {
			continue
		}

		re := regexp.MustCompile("(^|[^a-z0-9])" + regexp.QuoteMeta(normalized) + "([^a-z0-9]|$)")
		if re.MatchString(lowerName) {
			return SafeModeMatch{Pattern: normalized, Level: dangerLevelForPattern(normalized)}, true
		}
	}

	return SafeModeMatch{}, false
}

func effectiveProductionPatterns(patterns []string) []string {
	if len(patterns) == 0 {
		return DefaultProductionPatterns()
	}

	return patterns
}

func dangerLevelForPattern(pattern string) DangerLevel {
	if pattern == "prod" {
		return DangerLevelWarning
	}

	return DangerLevelDanger
}
