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
	if len(cfg.Favorites) != 0 {
		t.Fatalf("Load() favorites = %v, want empty", cfg.Favorites)
	}
}

func TestSaveAndLoadRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "config.yaml")
	t.Setenv(envConfigPath, path)

	want := Config{Favorites: []string{"prod-admin", "sandbox"}}
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
