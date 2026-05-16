package config

import (
	"os"
	"path/filepath"

	"fdroidadb/internal/xdg"
	"gopkg.in/yaml.v3"
)

type Repo struct {
	Name string `yaml:"name"`
	URL  string `yaml:"url"`
}

type Config struct {
	Repos      []Repo `yaml:"repos"`
	ADBPath    string `yaml:"adb_path"`
	MaxRetries int    `yaml:"max_retries"`
}

func DefaultConfig() *Config {
	return &Config{
		Repos: []Repo{
			{
				Name: "F-Droid",
				URL:  "https://f-droid.org/repo",
			},
		},
		ADBPath:    "adb",
		MaxRetries: 3,
	}
}

func Load() (*Config, error) {
	configDir := xdg.ConfigDir()
	configFile := filepath.Join(configDir, "config.yaml")

	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		cfg := DefaultConfig()
		if err := Save(cfg); err != nil {
			return nil, err
		}
		return cfg, nil
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func Save(cfg *Config) error {
	configDir := xdg.ConfigDir()
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}
	configFile := filepath.Join(configDir, "config.yaml")

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(configFile, data, 0644)
}
