package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestLoadMissingConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	t.Setenv(envConfigPath, path)

	cfg, gotPath, err := Load()
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}
	if gotPath != path {
		t.Fatalf("Load() path = %q, want %q", gotPath, path)
	}
	if len(cfg.Favorites) != 0 || len(cfg.Recents) != 0 {
		t.Fatalf("Load() config = %+v, want empty favorites and recents", cfg)
	}
	if !cfg.ConfirmProduction {
		t.Fatal("Load() ConfirmProduction = false, want true")
	}
	if !reflect.DeepEqual(cfg.ProductionPatterns, DefaultProductionPatterns()) {
		t.Fatalf("Load() ProductionPatterns = %+v, want %+v", cfg.ProductionPatterns, DefaultProductionPatterns())
	}
	if len(cfg.RiskPatterns) != 0 {
		t.Fatalf("Load() RiskPatterns = %+v, want empty", cfg.RiskPatterns)
	}
}

func TestSaveAndLoadRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "config.yaml")
	t.Setenv(envConfigPath, path)

	want := Config{
		Favorites:          []string{"prod-admin", "sandbox"},
		Recents:            []string{"sandbox", "dev"},
		ProductionPatterns: []string{"prod", "critical"},
		ConfirmProduction:  false,
		RiskPatterns:       []string{"prod", "critical"},
	}
	if err := Save(want, path); err != nil {
		t.Fatalf("Save() unexpected error: %v", err)
	}

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("config file not written: %v", err)
	}

	got, gotPath, err := Load()
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}
	if gotPath != path {
		t.Fatalf("Load() path = %q, want %q", gotPath, path)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Load() config = %+v, want %+v", got, want)
	}
}

func TestLoadRiskPatternsFallback(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	t.Setenv(envConfigPath, path)

	if err := os.WriteFile(path, []byte("risk_patterns:\n  - critical\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() unexpected error: %v", err)
	}

	got, _, err := Load()
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}

	if !reflect.DeepEqual(got.ProductionPatterns, []string{"critical"}) {
		t.Fatalf("Load() ProductionPatterns = %+v, want [critical]", got.ProductionPatterns)
	}
}
