package gitops

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetGitAuthor(t *testing.T) {
	// This test relies on git config being set
	author := GetGitAuthor()
	if author == "" {
		t.Skip("git config user.email not set")
	}

	if !contains(author, "@") {
		t.Errorf("expected email format with @, got: %s", author)
	}
}

func TestCalculateLOCFromRepos_EmptyRepos(t *testing.T) {
	additions, deletions, err := CalculateLOCFromRepos("2026-01-01", "2026-01-02", []string{}, "test@example.com")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if additions != 0 || deletions != 0 {
		t.Errorf("expected 0 additions and deletions for empty repos, got +%d -%d", additions, deletions)
	}
}

func TestCalculateLOCFromRepos_NonexistentRepo(t *testing.T) {
	nonexistentPath := filepath.Join(os.TempDir(), "nonexistent-repo-12345")

	// Should not error, just skip the repo
	additions, deletions, err := CalculateLOCFromRepos("2026-01-01", "2026-01-02", []string{nonexistentPath}, "test@example.com")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Should return 0 since repo doesn't exist
	if additions != 0 || deletions != 0 {
		t.Errorf("expected 0 additions/deletions for nonexistent repo, got +%d -%d", additions, deletions)
	}
}

func TestEnsureReposAvailable_Integration(t *testing.T) {
	// Skip in CI or if gh not available
	if os.Getenv("CI") == "true" {
		t.Skip("skipping integration test in CI")
	}

	// This is an integration test that requires gh CLI and network
	// Using a date range with no commits to avoid cloning
	repos, err := EnsureReposAvailable("2020-01-01", "2020-01-02")
	if err != nil {
		t.Logf("gh search might not be available: %v", err)
		return
	}

	// Should return empty list or valid repo paths
	for _, repo := range repos {
		if !filepath.IsAbs(repo) {
			t.Errorf("expected absolute path, got: %s", repo)
		}
	}
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && len(s) >= len(substr) &&
		(s == substr || len(s) > len(substr) && contains(s[1:], substr) || s[0:len(substr)] == substr)
}
