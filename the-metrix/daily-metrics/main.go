package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/sondy91/daily-metrics/internal/apiclient"
)

type JiraIssue struct {
	Type    string
	Key     string
	Summary string
	Status  string
}

type GitHubPR struct {
	Number     int    `json:"number"`
	Title      string `json:"title"`
	State      string `json:"state"`
	UpdatedAt  string `json:"updatedAt"`
	Repository struct {
		Name          string `json:"name"`
		NameWithOwner string `json:"nameWithOwner"`
	} `json:"repository"`
}

type DailyMetrics struct {
	Date         string
	JiraIssues   []JiraIssue
	GitHubPRs    []GitHubPR
	GitCommits   int
	LinesAdded   int
	LinesDeleted int
}

func main() {
	// Check for subcommands
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "record":
			handleRecord()
			return
		case "summary":
			handleSummary()
			return
		case "opencode":
			// Check for opencode subcommands
			if len(os.Args) > 2 && os.Args[2] == "submit" {
				handleOpenCodeSubmit()
				return
			}
			handleOpenCode()
			return
		}
	}

	// Parse flags for generate command
	allRepos := flag.Bool("all", false, "Include all repos (public and private)")
	outputFile := flag.String("output", "", "Output markdown to file (e.g., daily note path)")
	apiURL := flag.String("api-url", "http://localhost:8080", "Metrics API base URL")
	forceOffline := flag.Bool("force-offline", false, "Skip API and use direct queries")
	verbose := flag.Bool("verbose", false, "Show API request details")
	flag.Parse()

	date := time.Now().Format("2006-01-02")
	if flag.NArg() > 0 {
		date = flag.Arg(0)
	}

	// Load config
	cfg, err := loadConfig()
	if err != nil {
		fmt.Printf("❌ Failed to load config: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Fetching metrics for %s...\n", date)
	if *allRepos {
		fmt.Println("📂 Including all repos (public + private)")
	} else {
		fmt.Println("📂 Public/work repos only")
	}

	metrics := DailyMetrics{Date: date}
	usingAPI := false

	// Try API first unless force-offline
	if !*forceOffline {
		client, err := apiclient.NewClient(*apiURL, *verbose)
		if err != nil {
			fmt.Printf("⚠️  Invalid API URL: %v\n", err)
			fmt.Println("📍 Falling back to direct queries")
		} else if client.CheckHealth() {
			fmt.Println("🌐 Using metrics-api")
			apiMetrics, err := client.GetMetrics(date, *allRepos)
			if err == nil {
				metrics.JiraIssues = convertJiraIssues(apiMetrics.JiraIssues)
				metrics.GitHubPRs = convertGitHubPRs(apiMetrics.GitHubPRs)
				metrics.GitCommits = apiMetrics.GitCommits
				metrics.LinesAdded = apiMetrics.LinesAdded
				metrics.LinesDeleted = apiMetrics.LinesDeleted
				usingAPI = true
			} else {
				fmt.Printf("⚠️  API error: %v\n", err)
				fmt.Println("📍 Falling back to direct queries")
			}
		} else {
			fmt.Println("📍 API unavailable, using direct queries")
		}
	} else {
		fmt.Println("📍 Using direct queries (offline mode)")
	}

	// Fallback to direct queries if API not used
	if !usingAPI {
		// Query Jira
		fmt.Println("📊 Querying Jira...")
		metrics.JiraIssues = queryJira(date)

		// Query GitHub
		fmt.Println("🔀 Querying GitHub...")
		metrics.GitHubPRs = queryGitHub(date)

		// Query Git commits
		fmt.Println("💾 Analyzing Git commits...")
		metrics.GitCommits, metrics.LinesAdded, metrics.LinesDeleted = queryGitCommits(date, cfg, *allRepos)
	}
	fmt.Println()

	// Generate output
	if *outputFile != "" {
		// Ensure directory exists
		dir := filepath.Dir(*outputFile)
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Printf("❌ Failed to create directory: %v\n", err)
			os.Exit(1)
		}

		// Write markdown to file
		markdown := generateMarkdown(metrics)
		if err := os.WriteFile(*outputFile, []byte(markdown), 0644); err != nil {
			fmt.Printf("❌ Failed to write output file: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("\n✅ Markdown written to: %s\n", *outputFile)
	} else {
		// Display to terminal
		displayMetrics(metrics)
	}
}

func convertJiraIssues(apiIssues []apiclient.JiraIssue) []JiraIssue {
	issues := make([]JiraIssue, len(apiIssues))
	for i, api := range apiIssues {
		issues[i] = JiraIssue{
			Type:    api.Type,
			Key:     api.Key,
			Summary: api.Summary,
			Status:  api.Status,
		}
	}
	return issues
}

func convertGitHubPRs(apiPRs []apiclient.GitHubPR) []GitHubPR {
	prs := make([]GitHubPR, len(apiPRs))
	for i, api := range apiPRs {
		prs[i] = GitHubPR{
			Number:    api.Number,
			Title:     api.Title,
			State:     api.State,
			UpdatedAt: api.UpdatedAt,
		}
		prs[i].Repository.Name = api.Repository.Name
		prs[i].Repository.NameWithOwner = api.Repository.NameWithOwner
	}
	return prs
}

func queryJira(date string) []JiraIssue {
	cmd := exec.Command("jira", "issue", "list",
		"--assignee", getJiraUser(),
		"--updated-after", date,
		"--plain")

	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("⚠️  Jira query failed: %v\n", err)
		return nil
	}

	lines := strings.Split(string(output), "\n")
	issues := []JiraIssue{}

	for i, line := range lines {
		if i == 0 || strings.TrimSpace(line) == "" {
			continue // Skip header and empty lines
		}

		fields := strings.Fields(line)
		if len(fields) >= 4 {
			issue := JiraIssue{
				Type:    fields[0],
				Key:     fields[1],
				Status:  fields[len(fields)-1],
				Summary: strings.Join(fields[2:len(fields)-1], " "),
			}
			issues = append(issues, issue)
		}
	}

	return issues
}

func queryGitHub(date string) []GitHubPR {
	cmd := exec.Command("gh", "search", "prs",
		"--author", "@me",
		"--updated", date,
		"--json", "number,title,state,repository,updatedAt",
		"--limit", "50")

	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("⚠️  GitHub query failed: %v\n", err)
		return nil
	}

	var prs []GitHubPR
	if err := json.Unmarshal(output, &prs); err != nil {
		fmt.Printf("⚠️  Failed to parse GitHub response: %v\n", err)
		return nil
	}

	return prs
}

func queryGitCommits(date string, cfg *Config, includePrivate bool) (commits, added, deleted int) {
	author := getGitAuthor()

	// Determine which repos to scan
	repos := cfg.WorkRepos
	if includePrivate {
		repos = append(repos, cfg.PrivateRepos...)
	}

	for _, repoPath := range repos {
		// Check if repo exists
		if _, err := os.Stat(repoPath); os.IsNotExist(err) {
			fmt.Printf("⚠️  Skipping non-existent repo: %s\n", repoPath)
			continue
		}

		// Get commit count
		cmd := exec.Command("git", "log",
			fmt.Sprintf("--author=%s", author),
			fmt.Sprintf("--since=%s 00:00", date),
			fmt.Sprintf("--until=%s 23:59", date),
			"--oneline")
		cmd.Dir = repoPath

		output, err := cmd.Output()
		if err != nil {
			fmt.Printf("⚠️  Git log query failed for %s: %v\n", repoPath, err)
			continue
		}

		repoCommits := 0
		if strings.TrimSpace(string(output)) != "" {
			repoCommits = len(strings.Split(strings.TrimSpace(string(output)), "\n"))
		}
		commits += repoCommits

		// Get LOC stats
		cmd = exec.Command("git", "log",
			fmt.Sprintf("--author=%s", author),
			fmt.Sprintf("--since=%s 00:00", date),
			fmt.Sprintf("--until=%s 23:59", date),
			"--numstat",
			"--pretty=format:")
		cmd.Dir = repoPath

		output, err = cmd.Output()
		if err != nil {
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

		added += repoAdded
		deleted += repoDeleted

		if repoCommits > 0 {
			repoName := filepath.Base(repoPath)
			fmt.Printf("   ✓ %s: %d commits (+%d -%d)\n", repoName, repoCommits, repoAdded, repoDeleted)
		}
	}

	return commits, added, deleted
}

func getJiraUser() string {
	cmd := exec.Command("jira", "me")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

func getGitAuthor() string {
	cmd := exec.Command("git", "config", "user.email")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

func displayMetrics(m DailyMetrics) {
	fmt.Printf("\n╔════════════════════════════════════════════════════════════╗\n")
	fmt.Printf("║  Daily Metrics for %s                              ║\n", m.Date)
	fmt.Printf("╚════════════════════════════════════════════════════════════╝\n\n")

	// Git stats
	fmt.Printf("📊 Git Activity\n")
	fmt.Printf("   • %d commits\n", m.GitCommits)
	fmt.Printf("   • +%d -%d LOC\n\n", m.LinesAdded, m.LinesDeleted)

	// GitHub PRs
	fmt.Printf("🔀 GitHub Pull Requests (%d)\n", len(m.GitHubPRs))

	merged := 0
	open := 0
	closed := 0

	for _, pr := range m.GitHubPRs {
		switch pr.State {
		case "merged":
			merged++
		case "open":
			open++
		case "closed":
			closed++
		}
	}

	fmt.Printf("   • %d merged\n", merged)
	fmt.Printf("   • %d open\n", open)
	if closed > 0 {
		fmt.Printf("   • %d closed\n", closed)
	}

	if len(m.GitHubPRs) > 0 {
		fmt.Printf("\n   Recent PRs:\n")
		for i, pr := range m.GitHubPRs {
			if i >= 5 {
				break
			}
			stateIcon := "●"
			switch pr.State {
			case "merged":
				stateIcon = "✓"
			case "open":
				stateIcon = "○"
			}
			fmt.Printf("   %s #%d: %s\n", stateIcon, pr.Number, pr.Title)
			fmt.Printf("      %s\n", pr.Repository.NameWithOwner)
		}
	}

	// Jira issues
	fmt.Printf("\n✅ Jira Issues (%d)\n", len(m.JiraIssues))

	statusCount := make(map[string]int)
	for _, issue := range m.JiraIssues {
		statusCount[issue.Status]++
	}

	for status, count := range statusCount {
		fmt.Printf("   • %d in %s\n", count, status)
	}

	if len(m.JiraIssues) > 0 {
		fmt.Printf("\n   Recent Issues:\n")
		for i, issue := range m.JiraIssues {
			if i >= 5 {
				break
			}
			fmt.Printf("   • [%s] %s - %s\n", issue.Key, issue.Status, issue.Summary)
		}
	}
}

func handleOpenCode() {
	days := flag.Int("days", 7, "Number of days to analyze (default: 7)")
	startDate := flag.String("start", "", "Start date (YYYY-MM-DD)")
	endDate := flag.String("end", "", "End date (YYYY-MM-DD)")
	wpmOverride := flag.Int("wpm", 0, "Typing speed override (WPM)")
	format := flag.String("format", "terminal", "Output format: terminal, json, markdown")
	output := flag.String("output", "", "Output file path (required for json/markdown formats)")
	flag.CommandLine.Parse(os.Args[2:])

	cfg, err := loadConfig()
	if err != nil {
		fmt.Printf("❌ Failed to load config: %v\n", err)
		os.Exit(1)
	}

	typingSpeed := cfg.TypingSpeedWPM
	if typingSpeed == 0 {
		typingSpeed = 50
	}
	if *wpmOverride > 0 {
		typingSpeed = *wpmOverride
	}

	var since, until time.Time

	if *startDate != "" && *endDate != "" {
		since, err = time.Parse("2006-01-02", *startDate)
		if err != nil {
			fmt.Printf("❌ Invalid start date format (use YYYY-MM-DD): %v\n", err)
			os.Exit(1)
		}
		until, err = time.Parse("2006-01-02", *endDate)
		if err != nil {
			fmt.Printf("❌ Invalid end date format (use YYYY-MM-DD): %v\n", err)
			os.Exit(1)
		}
		until = until.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
	} else {
		until = time.Now()
		since = until.AddDate(0, 0, -*days)
	}

	if (*format == "json" || *format == "markdown") && *output == "" {
		fmt.Printf("❌ Output file required for %s format. Use --output <path>\n", *format)
		os.Exit(1)
	}

	if *format != "terminal" && *format != "json" && *format != "markdown" {
		fmt.Printf("❌ Invalid format: %s. Must be one of: terminal, json, markdown\n", *format)
		os.Exit(1)
	}

	if *format == "terminal" {
		fmt.Printf("🔍 Analyzing OpenCode usage from %s to %s...\n", since.Format("2006-01-02"), until.Format("2006-01-02"))
		fmt.Println()
	}

	stats, err := analyzeOpenCodeUsage(since, until, typingSpeed)
	if err != nil {
		fmt.Printf("❌ Error analyzing OpenCode data: %v\n", err)
		os.Exit(1)
	}

	if stats.TotalCommands == 0 {
		fmt.Println("📭 No OpenCode sessions found in this date range.")
		fmt.Println()
		fmt.Println("💡 Make sure you've used OpenCode and the database exists at:")
		fmt.Println("   ~/.local/share/opencode/opencode.db")
		return
	}

	switch *format {
	case "json":
		if err := exportJSON(stats, *output); err != nil {
			fmt.Printf("❌ Failed to export JSON: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✅ JSON exported to: %s\n", *output)
	case "markdown":
		if err := exportMarkdown(stats, typingSpeed, *output); err != nil {
			fmt.Printf("❌ Failed to export Markdown: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✅ Markdown exported to: %s\n", *output)
	default:
		formatOpenCodeStats(stats, typingSpeed, since, until)
	}
}
