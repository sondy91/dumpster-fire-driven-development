package gitops

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func EnsureReposAvailable(startDate, endDate string) ([]string, error) {
	repos, err := getReposInDateRange(startDate, endDate)
	if err != nil {
		return nil, err
	}

	var repoPaths []string
	projectsDir := filepath.Join(os.Getenv("HOME"), "projects")

	for _, repoFullName := range repos {
		parts := strings.Split(repoFullName, "/")
		if len(parts) != 2 {
			continue
		}
		repoName := parts[1]

		repoPath := filepath.Join(projectsDir, repoName)

		if _, err := os.Stat(repoPath); err == nil {
			repoPaths = append(repoPaths, repoPath)
			continue
		}

		fmt.Printf("   📦 Cloning %s to ~/projects/%s...\n", repoFullName, repoName)

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

func CalculateLOCFromRepos(startDate, endDate string, repoPaths []string, author string) (additions, deletions int, err error) {
	for _, repoPath := range repoPaths {
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
				if _, err := fmt.Sscanf(fields[0], "%d", &a); err != nil {
					continue
				}
				if _, err := fmt.Sscanf(fields[1], "%d", &d); err != nil {
					continue
				}
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

func GetGitAuthor() string {
	cmd := exec.Command("git", "config", "user.email")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}
