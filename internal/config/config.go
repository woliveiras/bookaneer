package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"gopkg.in/yaml.v3"
)

// Config holds all application configuration.
type Config struct {
	Port                  int                    `yaml:"port"`
	BindAddress           string                 `yaml:"bindAddress"`
	URLBase               string                 `yaml:"urlBase"`
	LogLevel              string                 `yaml:"logLevel"`
	DataDir               string                 `yaml:"-"`
	LibraryDir            string                 `yaml:"libraryDir"`
	AuthMethod            string                 `yaml:"authMethod"`
	CustomProviders       []CustomProviderConfig `yaml:"customProviders"`
	CustomProvidersEnable bool                   `yaml:"customProvidersEnabled"`
}

type CustomProviderConfig struct {
	Name       string `yaml:"name" json:"name"`
	Domain     string `yaml:"domain" json:"domain"`
	FormatHint string `yaml:"formatHint" json:"formatHint"`
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		Port:                  9090,
		BindAddress:           "0.0.0.0",
		URLBase:               "",
		LogLevel:              "info",
		AuthMethod:            "forms",
		LibraryDir:            "/library",
		CustomProvidersEnable: false,
	}
}

// Load reads configuration from environment variables and config file.
func Load(dataDir, configPath string) (*Config, error) {
	cfg := DefaultConfig()

	if dataDir == "" {
		dataDir = os.Getenv("BOOKANEER_DATA_DIR")
	}
	if dataDir == "" {
		dataDir = "./data"
	}
	cfg.DataDir = dataDir

	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return nil, fmt.Errorf("create data directory: %w", err)
	}

	if configPath == "" {
		configPath = filepath.Join(dataDir, "config.yaml")
	}

	if _, err := os.Stat(configPath); err == nil {
		data, err := os.ReadFile(configPath)
		if err != nil {
			return nil, fmt.Errorf("read config file: %w", err)
		}
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("parse config file: %w", err)
		}
	}

	if v := os.Getenv("BOOKANEER_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			cfg.Port = port
		}
	}
	if v := os.Getenv("BOOKANEER_BIND_ADDRESS"); v != "" {
		cfg.BindAddress = v
	}
	if v := os.Getenv("BOOKANEER_LOG_LEVEL"); v != "" {
		cfg.LogLevel = v
	}
	if v := os.Getenv("BOOKANEER_LIBRARY_DIR"); v != "" {
		cfg.LibraryDir = v
	}

	return cfg, nil
}

// DatabasePath returns the path to the SQLite database file.
func (c *Config) DatabasePath() string {
	return filepath.Join(c.DataDir, "bookaneer.db")
}
