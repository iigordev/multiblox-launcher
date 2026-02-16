package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Instance defines what we save in the JSON file
type Instance struct {
	Name     string `json:"name"`
	Path     string `json:"path"`
	IconName string `json:"icon_name"` // Store "testlogo1" (without extension)
}

// Config wraps the list of instances
type Config struct {
	Instances []Instance `json:"instances"`
}

// GetConfigPath points to ~/Library/Application Support/MultiBlox/config.json
func GetConfigPath() string {
	home, _ := os.UserHomeDir()
	path := filepath.Join(home, "Library", "Application Support", "MultiBlox", "config.json")
	return path
}

// Load reads the JSON file from disk
func Load() (*Config, error) {
	path := GetConfigPath()
	file, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{Instances: []Instance{}}, nil
		}
		return nil, err
	}

	var cfg Config
	err = json.Unmarshal(file, &cfg)
	return &cfg, err
}

// Save writes the current instance list to the JSON file
func Save(cfg *Config) error {
	path := GetConfigPath()
	// Create the directory if it's missing
	os.MkdirAll(filepath.Dir(path), 0755)

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
