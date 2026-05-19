package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	WorkRepos       []string `json:"work_repos"`
	PrivateRepos    []string `json:"private_repos,omitempty"`
	ScheduleBanchor string   `json:"schedule_b_anchor,omitempty"`
	Holidays2025    []string `json:"holidays_2025,omitempty"`
	Holidays2026    []string `json:"holidays_2026,omitempty"`
	JiraUser        string   `json:"jira_user,omitempty"`
	GitEmail        string   `json:"git_email,omitempty"`
	TypingSpeedWPM  int      `json:"typing_speed_wpm,omitempty"`
	MetricsAPIURL   string   `json:"metrics_api_url,omitempty"`
}

func loadConfig() (*Config, error) {
	configPath := filepath.Join(os.Getenv("HOME"), ".config", "daily-metrics", "config.json")

	// Create default config if doesn't exist
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Use environment variable or default project name
		projectName := os.Getenv("PROJECT_NAME")
		if projectName == "" {
			projectName = "my-project"
		}
		
		defaultConfig := &Config{
			WorkRepos: []string{
				filepath.Join(os.Getenv("HOME"), "projects", projectName),
			},
			PrivateRepos: []string{},
			TypingSpeedWPM: 50,
			MetricsAPIURL:  "http://localhost:8080",
		}

		if err := saveConfig(defaultConfig); err != nil {
			return nil, fmt.Errorf("failed to create default config: %w", err)
		}

		fmt.Printf("✨ Created default config at %s\n", configPath)
		fmt.Println("   Edit this file to customize tracked repositories")

		return defaultConfig, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &cfg, nil
}

func saveConfig(cfg *Config) error {
	configDir := filepath.Join(os.Getenv("HOME"), ".config", "daily-metrics")

	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	configPath := filepath.Join(configDir, "config.json")

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}
