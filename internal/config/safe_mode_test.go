package config

import "testing"

func TestDefaultProductionPatterns(t *testing.T) {
	patterns := DefaultProductionPatterns()
	if len(patterns) != 3 {
		t.Fatalf("DefaultProductionPatterns() len = %d, want 3", len(patterns))
	}

	patterns[0] = "changed"
	if DefaultProductionPatterns()[0] != "prod" {
		t.Fatal("DefaultProductionPatterns() should return a copy")
	}
}

func TestMatchProductionProfile(t *testing.T) {
	tests := []struct {
		name       string
		profile    string
		patterns   []string
		wantMatch  bool
		wantLevel  DangerLevel
		wantPattern string
	}{
		{name: "default prod match", profile: "prod-admin", wantMatch: true, wantLevel: DangerLevelWarning, wantPattern: "prod"},
		{name: "production match", profile: "team-production", wantMatch: true, wantLevel: DangerLevelDanger, wantPattern: "production"},
		{name: "live match", profile: "live_eu", wantMatch: true, wantLevel: DangerLevelDanger, wantPattern: "live"},
		{name: "custom pattern", profile: "critical-admin", patterns: []string{"critical"}, wantMatch: true, wantLevel: DangerLevelDanger, wantPattern: "critical"},
		{name: "does not match product", profile: "product-team", wantMatch: false},
		{name: "does not match empty", profile: "sandbox", patterns: []string{"", "   "}, wantMatch: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			match, ok := MatchProductionProfile(tt.profile, tt.patterns)
			if ok != tt.wantMatch {
				t.Fatalf("MatchProductionProfile() matched = %v, want %v", ok, tt.wantMatch)
			}
			if !tt.wantMatch {
				return
			}

			if match.Level != tt.wantLevel {
				t.Fatalf("MatchProductionProfile() level = %q, want %q", match.Level, tt.wantLevel)
			}
			if match.Pattern != tt.wantPattern {
				t.Fatalf("MatchProductionProfile() pattern = %q, want %q", match.Pattern, tt.wantPattern)
			}
		})
	}
}
