package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `repos:
  work:
    - main-project
    - test-repo
  private:
    - SecondBrain
schedule_b_anchor: "2026-04-03"
holidays_2026:
  - "2026-12-25"
  - "2026-07-04"
users:
  - e_number: e12345678
    jira_account_id: "712020:abcd1234"
    github_username: testuser
    name: Test User
    email: test@example.com
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if len(cfg.Repos.Work) != 2 {
		t.Errorf("expected 2 work repos, got %d", len(cfg.Repos.Work))
	}

	if cfg.Repos.Work[1] != "test-repo" {
		t.Errorf("expected repo name test-repo, got %s", cfg.Repos.Work[1])
	}

	if cfg.ScheduleBanchor != "2026-04-03" {
		t.Errorf("expected schedule_b_anchor 2026-04-03, got %s", cfg.ScheduleBanchor)
	}

	if len(cfg.Holidays2026) != 2 {
		t.Errorf("expected 2 holidays, got %d", len(cfg.Holidays2026))
	}

	if len(cfg.Users) != 1 {
		t.Errorf("expected 1 user, got %d", len(cfg.Users))
	}

	if cfg.Users[0].ENumber != "e12345678" {
		t.Errorf("expected e_number e12345678, got %s", cfg.Users[0].ENumber)
	}
}

func TestSaveConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	cfg := &MetricsConfig{
		Repos: ReposSection{
			Work:    []string{"main-project", "test-repo"},
			Private: []string{"SecondBrain"},
		},
		ScheduleBanchor: "2026-04-03",
		Holidays2026:    []string{"2026-12-25"},
		Users: []UserConfig{
			{
				ENumber:        "e12345678",
				JiraAccountID:  "712020:abcd1234",
				GitHubUsername: "testuser",
				Name:           "Test User",
				Email:          "test@example.com",
			},
		},
	}

	err := Save(cfg, configPath)
	if err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("config file was not created")
	}

	loadedCfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("failed to load saved config: %v", err)
	}

	if len(loadedCfg.Repos.Work) != 2 {
		t.Errorf("expected 2 work repos after save/load, got %d", len(loadedCfg.Repos.Work))
	}

	if loadedCfg.Users[0].ENumber != "e12345678" {
		t.Errorf("expected e_number e12345678 after save/load, got %s", loadedCfg.Users[0].ENumber)
	}
}

func TestLoadConfigNotFound(t *testing.T) {
	_, err := Load("/nonexistent/path/config.yaml")
	if err == nil {
		t.Error("expected error when loading nonexistent config, got nil")
	}
}
