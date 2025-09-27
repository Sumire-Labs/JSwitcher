package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	CustomSearchPaths []string `json:"custom_search_paths"`
	Theme             string   `json:"theme"`
	AutoDetect        bool     `json:"auto_detect"`
}

func DefaultConfig() *Config {
	return &Config{
		CustomSearchPaths: []string{},
		Theme:             "default",
		AutoDetect:        true,
	}
}

func (c *Config) Save(configPath string) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	// Ensure directory exists
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

func LoadConfig(configPath string) (*Config, error) {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Return default config if file doesn't exist
		return DefaultConfig(), nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func GetConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "./javaswitcher.json"
	}
	return filepath.Join(homeDir, ".javaswitcher", "config.json")
}