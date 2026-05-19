package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ensureReposAvailable clones any missing repos needed for LOC calculation
// Returns list of all repo paths (existing + newly cloned)
func ensureReposAvailable(startDate, endDate string) ([]string, error) {
	// Get all repos committed to in date range
	repos, err := getReposInDateRange(startDate, endDate)
	if err != nil {
		return nil, err
	}

	var repoPaths []string
	projectsDir := filepath.Join(os.Getenv("HOME"), "projects")

	for _, repoFullName := range repos {
		// Extract repo name from fullName (e.g., "myorg/platform-ops" -> "plat-gitops-npd")
		parts := strings.Split(repoFullName, "/")
		if len(parts) != 2 {
			continue
		}
		repoName := parts[1]

		// Check if repo exists locally
		repoPath := filepath.Join(projectsDir, repoName)

		if _, err := os.Stat(repoPath); err == nil {
			// Repo exists locally
			repoPaths = append(repoPaths, repoPath)
			continue
		}

		// Repo doesn't exist, clone it
		fmt.Printf("   📦 Cloning %s to ~/projects/%s...\n", repoFullName, repoName)

		// Use GitHub CLI to clone (handles authentication)
		cmd := exec.Command("gh", "repo", "clone", repoFullName, repoPath)
		output, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("   ⚠️  Failed to clone %s: %v\n%s\n", repoFullName, err, string(output))
			continue
		}

		fmt.Printf("   ✓ Cloned %s\n", repoName)
		repoPaths = append(repoPaths, repoPath)
	}

	return repoPaths, nil
}

// getReposInDateRange returns list of repos committed to in date range
func getReposInDateRange(startDate, endDate string) ([]string, error) {
	cmd := exec.Command("gh", "search", "commits",
		"--author", "@me",
		fmt.Sprintf("--author-date=%s..%s", startDate, endDate),
		"--json", "repository",
		"--limit", "1000")

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to query commits: %w", err)
	}

	var results []struct {
		Repository struct {
			FullName string `json:"fullName"`
		} `json:"repository"`
	}

	if err := json.Unmarshal(output, &results); err != nil {
		return nil, fmt.Errorf("failed to parse commit results: %w", err)
	}

	// Get unique repo names
	repoSet := make(map[string]bool)
	for _, result := range results {
		if result.Repository.FullName != "" {
			repoSet[result.Repository.FullName] = true
		}
	}

	var repos []string
	for repo := range repoSet {
		repos = append(repos, repo)
	}

	return repos, nil
}

// querySummaryLocalLOCFromAllRepos calculates LOC from all repos (clones missing ones)
func querySummaryLocalLOCFromAllRepos(startDate, endDate string, cfg *Config, includePrivate bool) (additions, deletions int, err error) {
	fmt.Println("   Ensuring all repos are available locally...")
	repoPaths, err := ensureReposAvailable(startDate, endDate)
	if err != nil {
		return 0, 0, err
	}

	if len(repoPaths) == 0 {
		return 0, 0, nil
	}

	// Build set of private repo names for filtering
	privateRepoNames := make(map[string]bool)
	if !includePrivate {
		for _, repoPath := range cfg.PrivateRepos {
			repoName := filepath.Base(repoPath)
			privateRepoNames[repoName] = true
		}
	}

	// Filter out private repos if needed
	var filteredRepoPaths []string
	for _, repoPath := range repoPaths {
		repoName := filepath.Base(repoPath)
		if !includePrivate && privateRepoNames[repoName] {
			continue
		}
		filteredRepoPaths = append(filteredRepoPaths, repoPath)
	}

	if len(filteredRepoPaths) == 0 {
		return 0, 0, nil
	}

	fmt.Printf("   Analyzing LOC from %d repos...\n", len(filteredRepoPaths))

	author := getGitAuthor()

	for _, repoPath := range filteredRepoPaths {
		cmd := exec.Command("git", "log",
			fmt.Sprintf("--author=%s", author),
			fmt.Sprintf("--since=%s 00:00", startDate),
			fmt.Sprintf("--until=%s 23:59", endDate),
			"--numstat",
			"--pretty=format:")
		cmd.Dir = repoPath

		output, err := cmd.Output()
		if err != nil {
			fmt.Printf("   ⚠️  Failed to get LOC for %s: %v\n", filepath.Base(repoPath), err)
			continue
		}

		repoAdded := 0
		repoDeleted := 0

		for _, line := range strings.Split(string(output), "\n") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				var a, d int
				fmt.Sscanf(fields[0], "%d", &a)
				fmt.Sscanf(fields[1], "%d", &d)
				repoAdded += a
				repoDeleted += d
			}
		}

		additions += repoAdded
		deletions += repoDeleted

		if repoAdded > 0 || repoDeleted > 0 {
			fmt.Printf("     ✓ %s: +%d -%d\n", filepath.Base(repoPath), repoAdded, repoDeleted)
		}
	}

	return additions, deletions, nil
}
