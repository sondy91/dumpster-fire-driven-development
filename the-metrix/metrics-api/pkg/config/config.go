package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type ReposSection struct {
	Work    []string `yaml:"work"`
	Private []string `yaml:"private,omitempty"`
}

type MetricsConfig struct {
	Repos           ReposSection `yaml:"repos"`
	ScheduleBanchor string       `yaml:"schedule_b_anchor,omitempty"`
	Holidays2025    []string     `yaml:"holidays_2025,omitempty"`
	Holidays2026    []string     `yaml:"holidays_2026,omitempty"`
	Users           []UserConfig `yaml:"users,omitempty"`
	Cache           CacheConfig  `yaml:"cache,omitempty"`
}

type CacheConfig struct {
	Enabled    bool `yaml:"enabled"`
	TTLSeconds int  `yaml:"ttl_seconds"`
}

type RepoConfig struct {
	Name string `yaml:"name"`
	Path string `yaml:"path"`
}

type UserConfig struct {
	ENumber        string `yaml:"e_number"`
	JiraAccountID  string `yaml:"jira_account_id,omitempty"`
	GitHubUsername string `yaml:"github_username,omitempty"`
	Name           string `yaml:"name,omitempty"`
	Email          string `yaml:"email,omitempty"`
}

func Load(configPath string) (*MetricsConfig, error) {
	if configPath == "" {
		configPath = "./config.yaml"
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file not found: %s", configPath)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg MetricsConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &cfg, nil
}

func Save(cfg *MetricsConfig, configPath string) error {
	if configPath == "" {
		configPath = "./config.yaml"
	}

	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}
