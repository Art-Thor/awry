package ui

import "strings"

var defaultRiskPatterns = []string{"prod", "production", "live"}

func (m Model) riskPatterns() []string {
	if len(m.configRiskPatterns) > 0 {
		return m.configRiskPatterns
	}
	return defaultRiskPatterns
}

func (m Model) isRiskyProfile(name string) bool {
	lower := strings.ToLower(name)
	for _, pattern := range m.riskPatterns() {
		if pattern == "" {
			continue
		}
		if strings.Contains(lower, strings.ToLower(pattern)) {
			return true
		}
	}
	return false
}

func (m Model) riskLabel(name string) string {
	if m.isRiskyProfile(name) {
		return "Production-like"
	}
	return "Standard"
}
