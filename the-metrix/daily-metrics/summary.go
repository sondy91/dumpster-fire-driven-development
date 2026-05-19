package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func handleSummary() {
	summaryFlags := flag.NewFlagSet("summary", flag.ExitOnError)
	startDate := summaryFlags.String("start", "", "Start date (YYYY-MM-DD)")
	endDate := summaryFlags.String("end", "", "End date (YYYY-MM-DD)")
	outputFile := summaryFlags.String("output", "", "Output markdown to file")
	allRepos := summaryFlags.Bool("all", false, "Include all repos (work + private)")

	summaryFlags.Parse(os.Args[2:])

	if *startDate == "" || *endDate == "" {
		fmt.Println("❌ Start and end dates required. Usage: daily-metrics summary --start 2025-10-13 --end 2026-04-13")
		os.Exit(1)
	}

	// Validate date format
	if _, err := time.Parse("2006-01-02", *startDate); err != nil {
		fmt.Printf("❌ Invalid start date format. Use YYYY-MM-DD\n")
		os.Exit(1)
	}
	if _, err := time.Parse("2006-01-02", *endDate); err != nil {
		fmt.Printf("❌ Invalid end date format. Use YYYY-MM-DD\n")
		os.Exit(1)
	}

	fmt.Printf("📊 Generating summary from %s to %s...\n", *startDate, *endDate)

	// Load config for local repos
	cfg, err := loadConfig()
	if err != nil {
		fmt.Printf("❌ Failed to load config: %v\n", err)
		os.Exit(1)
	}

	if *allRepos {
		fmt.Println("📂 Including all repos (work + private)")
	} else {
		fmt.Println("📂 Work repos only")
	}
	fmt.Println()

	// Query Git stats from GitHub API (covers ALL repos you contributed to)
	fmt.Println("💾 Querying GitHub commits across all repos...")
	commits := querySummaryGitHubCommits(*startDate, *endDate, cfg, *allRepos)

	// Query LOC from all git repos (auto-clone missing ones)
	fmt.Println("📊 Analyzing LOC from all repos...")
	additions, deletions, err := querySummaryLocalLOCFromAllRepos(*startDate, *endDate, cfg, *allRepos)
	if err != nil {
		fmt.Printf("⚠️  LOC calculation failed: %v\n", err)
		additions, deletions = 0, 0
	}

	// Query GitHub stats
	fmt.Println("🔀 Querying GitHub PRs...")
	totalPRs, mergedPRs := querySummaryGitHub(*startDate, *endDate)

	// Query Jira stats
	fmt.Println("✅ Querying Jira issues...")
	totalJira, doneJira := querySummaryJira(*startDate, *endDate)

	// Query OpenCode stats
	fmt.Println("🤖 Querying OpenCode AI usage...")
	ocCommands, ocEdits, ocTimeSaved, ocErr := getOpenCodeSummaryStats(*startDate, *endDate)
	if ocErr != nil {
		if errors.Is(ocErr, ErrNoData) {
			fmt.Printf("ℹ️  No OpenCode usage in date range\n")
		} else {
			fmt.Printf("⚠️  OpenCode stats unavailable: %v\n", ocErr)
		}
		ocCommands, ocEdits, ocTimeSaved = 0, 0, 0
	}

	// Generate output
	if *outputFile != "" {
		generateSummaryMarkdown(*outputFile, *startDate, *endDate, commits, additions, deletions, totalPRs, mergedPRs, totalJira, doneJira, ocCommands, ocEdits, ocTimeSaved, cfg)
	} else {
		displaySummary(*startDate, *endDate, commits, additions, deletions, totalPRs, mergedPRs, totalJira, doneJira, ocCommands, ocEdits, ocTimeSaved, cfg)
	}
}

func querySummaryGitHubCommits(startDate, endDate string, cfg *Config, includePrivate bool) int {
	// Use GitHub search API to find all commits by author in date range
	cmd := exec.Command("gh", "search", "commits",
		"--author", "@me",
		fmt.Sprintf("--author-date=%s..%s", startDate, endDate),
		"--json", "sha,repository",
		"--limit", "1000")

	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("⚠️  GitHub commits query failed: %v\n", err)
		return 0
	}

	var commitResults []struct {
		SHA        string `json:"sha"`
		Repository struct {
			FullName string `json:"fullName"`
		} `json:"repository"`
	}

	if err := json.Unmarshal(output, &commitResults); err != nil {
		fmt.Printf("⚠️  Failed to parse GitHub commits response: %v\n", err)
		return 0
	}

	// Build set of private repo names for filtering
	privateRepoNames := make(map[string]bool)
	if !includePrivate {
		for _, repoPath := range cfg.PrivateRepos {
			repoName := filepath.Base(repoPath)
			privateRepoNames[repoName] = true
		}
	}

	// Count commits per repo (filter private if needed)
	repoCommits := make(map[string]int)
	filteredCount := 0
	for _, commit := range commitResults {
		repoFullName := commit.Repository.FullName
		if repoFullName == "" {
			continue
		}

		// Extract repo name from fullName
		parts := strings.Split(repoFullName, "/")
		if len(parts) != 2 {
			continue
		}
		repoName := parts[1]

		// Skip private repos if not including them
		if !includePrivate && privateRepoNames[repoName] {
			continue
		}

		repoCommits[repoFullName]++
		filteredCount++
	}

	// Display top repos sorted by commit count
	if len(repoCommits) > 0 {
		type repoStat struct {
			name    string
			commits int
		}
		var repos []repoStat
		for repo, commits := range repoCommits {
			repos = append(repos, repoStat{name: repo, commits: commits})
		}

		// Sort by commit count descending
		for i := 0; i < len(repos); i++ {
			for j := i + 1; j < len(repos); j++ {
				if repos[j].commits > repos[i].commits {
					repos[i], repos[j] = repos[j], repos[i]
				}
			}
		}

		fmt.Printf("   Top repos by commits:\n")
		for i, repo := range repos {
			if i >= 5 {
				break
			}
			fmt.Printf("     • %s: %d commits\n", repo.name, repo.commits)
		}
	}

	return filteredCount
}

func querySummaryLocalLOC(startDate, endDate string, cfg *Config) (additions, deletions int) {
	author := getGitAuthor()

	for _, repoPath := range cfg.WorkRepos {
		if _, err := os.Stat(repoPath); os.IsNotExist(err) {
			continue
		}

		cmd := exec.Command("git", "log",
			fmt.Sprintf("--author=%s", author),
			fmt.Sprintf("--since=%s 00:00", startDate),
			fmt.Sprintf("--until=%s 23:59", endDate),
			"--numstat",
			"--pretty=format:")
		cmd.Dir = repoPath

		output, err := cmd.Output()
		if err != nil {
			continue
		}

		for _, line := range strings.Split(string(output), "\n") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				var a, d int
				fmt.Sscanf(fields[0], "%d", &a)
				fmt.Sscanf(fields[1], "%d", &d)
				additions += a
				deletions += d
			}
		}
	}

	return additions, deletions
}

func querySummaryGitHub(startDate, endDate string) (total, merged int) {
	// Query all PRs created in date range (--merged date filter doesn't work reliably)
	cmd := exec.Command("gh", "search", "prs",
		"--author", "@me",
		"--created", fmt.Sprintf("%s..%s", startDate, endDate),
		"--json", "number,state",
		"--limit", "1000")

	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("⚠️  GitHub PR query failed: %v\n", err)
		return 0, 0
	}

	var prs []struct {
		Number int    `json:"number"`
		State  string `json:"state"`
	}

	if err := json.Unmarshal(output, &prs); err != nil {
		return 0, 0
	}

	total = len(prs)

	// Count merged PRs
	for _, pr := range prs {
		if pr.State == "merged" || pr.State == "MERGED" {
			merged++
		}
	}

	return total, merged
}

func querySummaryJira(startDate, endDate string) (total, done int) {
	cmd := exec.Command("jira", "issue", "list",
		"--assignee", getJiraUser(),
		"--updated-after", startDate,
		"--updated-before", endDate,
		"--plain")

	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("⚠️  Jira query failed: %v\n", err)
		return 0, 0
	}

	lines := strings.Split(string(output), "\n")

	for i, line := range lines {
		if i == 0 || strings.TrimSpace(line) == "" {
			continue
		}

		total++

		if strings.Contains(line, "Done") {
			done++
		}
	}

	return total, done
}

func displaySummary(startDate, endDate string, commits, additions, deletions, totalPRs, mergedPRs, totalJira, doneJira, ocCommands, ocEdits int, ocTimeSaved float64, cfg *Config) {
	fmt.Printf("\n╔════════════════════════════════════════════════════════════╗\n")
	fmt.Printf("║  Summary: %s to %s                  ║\n", startDate, endDate)
	fmt.Printf("╚════════════════════════════════════════════════════════════╝\n\n")

	// Git stats
	fmt.Printf("💾 GitHub Commits (across all repos)\n")
	fmt.Printf("   • %d total commits\n", commits)
	fmt.Printf("   • +%d -%d LOC\n", additions, deletions)
	fmt.Printf("   • Net change: %+d lines\n\n", additions-deletions)

	// GitHub stats
	fmt.Printf("🔀 GitHub Pull Requests\n")
	if totalPRs > 0 {
		fmt.Printf("   • %d total PRs merged\n", totalPRs)
		fmt.Printf("   • %d merged (%.1f%%)\n\n", mergedPRs, float64(mergedPRs)/float64(totalPRs)*100)
	} else {
		fmt.Printf("   • 0 PRs merged in date range\n\n")
	}

	// Jira stats
	fmt.Printf("✅ Jira Issues\n")
	if totalJira > 0 {
		fmt.Printf("   • %d total issues worked on\n", totalJira)
		fmt.Printf("   • %d completed (%.1f%%)\n\n", doneJira, float64(doneJira)/float64(totalJira)*100)
	} else {
		fmt.Printf("   • 0 issues worked on\n\n")
	}

	// OpenCode stats
	if ocCommands > 0 || ocEdits > 0 {
		fmt.Printf("🤖 OpenCode AI Assistance\n")
		fmt.Printf("   • %d bash commands generated\n", ocCommands)
		fmt.Printf("   • %d file edits\n", ocEdits)
		fmt.Printf("   • ~%.1f hours saved\n\n", ocTimeSaved/60.0)
	}

	// Productivity metrics with working days
	calendarDays := calculateDays(startDate, endDate)
	workingDays := calculateWorkingDays(startDate, endDate, cfg)
	weeks := float64(calendarDays) / 7.0

	fmt.Printf("📈 Productivity Metrics\n")
	fmt.Printf("   📅 %d calendar days (%d working days, %.1f weeks)\n", calendarDays, workingDays, weeks)

	if workingDays > 0 {
		fmt.Printf("   💻 %.1f commits/working day\n", float64(commits)/float64(workingDays))
		fmt.Printf("   📝 %.1f LOC/working day\n", float64(additions-deletions)/float64(workingDays))
	}

	fmt.Printf("   🔀 %.1f PRs/week\n", float64(totalPRs)/weeks)
	fmt.Printf("   ✅ %.1f issues completed/week\n", float64(doneJira)/weeks)
}

func calculateDays(startDate, endDate string) int {
	start, _ := time.Parse("2006-01-02", startDate)
	end, _ := time.Parse("2006-01-02", endDate)
	return int(end.Sub(start).Hours() / 24)
}

func generateSummaryMarkdown(outputPath, startDate, endDate string, commits, additions, deletions, totalPRs, mergedPRs, totalJira, doneJira, ocCommands, ocEdits int, ocTimeSaved float64, cfg *Config) {
	calendarDays := calculateDays(startDate, endDate)
	workingDays := calculateWorkingDays(startDate, endDate, cfg)
	weeks := float64(calendarDays) / 7.0

	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# Summary: %s to %s\n\n", startDate, endDate))

	sb.WriteString("## Git Activity\n\n")
	sb.WriteString(fmt.Sprintf("- **%d** total commits across all repos\n", commits))
	sb.WriteString(fmt.Sprintf("- **+%d -%d** LOC\n", additions, deletions))
	sb.WriteString(fmt.Sprintf("- Net change: **%+d** lines\n\n", additions-deletions))

	sb.WriteString("## GitHub Pull Requests\n\n")
	if totalPRs > 0 {
		sb.WriteString(fmt.Sprintf("- **%d** total PRs\n", totalPRs))
		sb.WriteString(fmt.Sprintf("- **%d** merged (%.1f%%)\n\n", mergedPRs, float64(mergedPRs)/float64(totalPRs)*100))
	} else {
		sb.WriteString("- 0 PRs in date range\n\n")
	}

	sb.WriteString("## Jira Issues\n\n")
	if totalJira > 0 {
		sb.WriteString(fmt.Sprintf("- **%d** total issues worked on\n", totalJira))
		sb.WriteString(fmt.Sprintf("- **%d** completed (%.1f%%)\n\n", doneJira, float64(doneJira)/float64(totalJira)*100))
	} else {
		sb.WriteString("- 0 issues worked on\n\n")
	}

	if ocCommands > 0 || ocEdits > 0 {
		sb.WriteString("## OpenCode AI Assistance\n\n")
		sb.WriteString(fmt.Sprintf("- **%d** bash commands generated\n", ocCommands))
		sb.WriteString(fmt.Sprintf("- **%d** file edits\n", ocEdits))
		sb.WriteString(fmt.Sprintf("- **~%.1f hours** saved\n\n", ocTimeSaved/60.0))
	}

	sb.WriteString("## Productivity Metrics\n\n")
	sb.WriteString(fmt.Sprintf("- **%d** calendar days (**%d** working days, **%.1f** weeks)\n", calendarDays, workingDays, weeks))

	if workingDays > 0 {
		sb.WriteString(fmt.Sprintf("- **%.1f** commits/working day\n", float64(commits)/float64(workingDays)))
		sb.WriteString(fmt.Sprintf("- **%.1f** LOC/working day\n", float64(additions-deletions)/float64(workingDays)))
	}

	sb.WriteString(fmt.Sprintf("- **%.1f** PRs/week\n", float64(totalPRs)/weeks))
	sb.WriteString(fmt.Sprintf("- **%.1f** issues completed/week\n", float64(doneJira)/weeks))

	if err := os.WriteFile(outputPath, []byte(sb.String()), 0644); err != nil {
		fmt.Printf("❌ Failed to write output file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n✅ Markdown written to: %s\n", outputPath)
}
