package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

const envConfigPath = "AWRY_CONFIG_PATH"

type Config struct {
	Favorites    []string `mapstructure:"favorites"`
	Recents      []string `mapstructure:"recents"`
	RiskPatterns []string `mapstructure:"risk_patterns"`
}

func Load() (Config, string, error) {
	path, err := configPath()
	if err != nil {
		return Config{}, "", err
	}

	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml")

	if err := v.ReadInConfig(); err != nil {
		var notFound viper.ConfigFileNotFoundError
		if !errors.As(err, &notFound) && !os.IsNotExist(err) {
			return Config{}, "", fmt.Errorf("reading config: %w", err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return Config{}, "", fmt.Errorf("decoding config: %w", err)
	}

	return cfg, path, nil
}

func Save(cfg Config, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml")
	v.Set("favorites", cfg.Favorites)
	v.Set("recents", cfg.Recents)
	v.Set("risk_patterns", cfg.RiskPatterns)

	if err := v.WriteConfigAs(path); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}

	return nil
}

func configPath() (string, error) {
	if override := os.Getenv(envConfigPath); override != "" {
		return override, nil
	}

	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("locating user config dir: %w", err)
	}

	return filepath.Join(configDir, "awry", "config.yaml"), nil
}
