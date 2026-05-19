package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type ActivitySubmission struct {
	ID                 string   `json:"id"`
	Date               string   `json:"date"`
	Developer          string   `json:"developer"`
	ProjectPath        string   `json:"project_path"`
	ProjectName        *string  `json:"project_name,omitempty"`
	BashCommands       int      `json:"bash_commands"`
	FileReads          int      `json:"file_reads"`
	FileEdits          int      `json:"file_edits"`
	FileWrites         int      `json:"file_writes"`
	Searches           int      `json:"searches"`
	TimeSavedBash      float64  `json:"time_saved_bash"`
	TimeSavedReads     float64  `json:"time_saved_reads"`
	TimeSavedEdits     float64  `json:"time_saved_edits"`
	TimeSavedWrites    float64  `json:"time_saved_writes"`
	TimeSavedSearches  float64  `json:"time_saved_searches"`
	TimeSavedTotal     float64  `json:"time_saved_total"`
	AvgComplexityScore float64  `json:"avg_complexity_score"`
	SessionCount       int      `json:"session_count"`
	CharsTyped         int      `json:"chars_typed"`
}

func handleOpenCodeSubmit() {
	submitCmd := flag.NewFlagSet("opencode submit", flag.ExitOnError)
	startDate := submitCmd.String("start", "", "Start date (YYYY-MM-DD), defaults to today")
	endDate := submitCmd.String("end", "", "End date (YYYY-MM-DD), defaults to today")
	developer := submitCmd.String("developer", "", "Developer name (defaults to git user.name)")
	apiURL := submitCmd.String("api-url", "", "Metrics API URL (defaults to config or http://localhost:8080)")
	dryRun := submitCmd.Bool("dry-run", false, "Print what would be submitted without actually sending")
	backfillAll := submitCmd.Bool("backfill-all", false, "Submit all historical OpenCode data (auto-detects earliest date)")

	submitCmd.Parse(os.Args[3:])

	cfg, err := loadConfig()
	if err != nil {
		fmt.Printf("❌ Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Determine API URL
	baseURL := *apiURL
	if baseURL == "" {
		baseURL = cfg.MetricsAPIURL
	}
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	// Determine date range (default to today, or last 90 days if backfill-all)
	var since, until time.Time
	now := time.Now()

	if *backfillAll {
		// Find earliest OpenCode data
		earliestDate, err := findEarliestOpenCodeData()
		if err != nil {
			fmt.Printf("⚠️ Could not determine earliest data, using last 365 days: %v\n", err)
			since = now.AddDate(0, 0, -365)
		} else {
			since = earliestDate
			fmt.Printf("🔄 Backfill mode: Found data going back to %s\n", earliestDate.Format("2006-01-02"))
		}
		until = now
		fmt.Printf("📅 Submitting all OpenCode data from %s to %s\n", since.Format("2006-01-02"), until.Format("2006-01-02"))
	} else if *startDate != "" {
		since, err = time.Parse("2006-01-02", *startDate)
		if err != nil {
			fmt.Printf("❌ Invalid start date format (use YYYY-MM-DD): %v\n", err)
			os.Exit(1)
		}
	} else {
		since = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	}

	if *endDate != "" && !*backfillAll {
		until, err = time.Parse("2006-01-02", *endDate)
		if err != nil {
			fmt.Printf("❌ Invalid end date format (use YYYY-MM-DD): %v\n", err)
			os.Exit(1)
		}
		until = until.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
	} else if !*backfillAll {
		until = time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 999999999, time.Local)
	}

	// Get developer name
	devName := *developer
	if devName == "" {
		devName = cfg.JiraUser
	}
	if devName == "" {
		// Try to get from git config
		devName = os.Getenv("USER")
	}
	if devName == "" {
		fmt.Printf("❌ Developer name required. Use --developer flag or set jira_user in config\n")
		os.Exit(1)
	}

	typingSpeed := cfg.TypingSpeedWPM
	if typingSpeed == 0 {
		typingSpeed = 50
	}

	fmt.Printf("📊 Collecting OpenCode activity from %s to %s...\n\n", since.Format("2006-01-02"), until.Format("2006-01-02"))

	// Iterate through each day in the date range
	var allSubmissions []ActivitySubmission
	daysWithActivity := 0
	for currentDate := since; !currentDate.After(until); currentDate = currentDate.AddDate(0, 0, 1) {
		dayStart := time.Date(currentDate.Year(), currentDate.Month(), currentDate.Day(), 0, 0, 0, 0, time.Local)
		dayEnd := time.Date(currentDate.Year(), currentDate.Month(), currentDate.Day(), 23, 59, 59, 999999999, time.Local)

		// Analyze this specific day
		dayStats, err := analyzeOpenCodeUsage(dayStart, dayEnd, typingSpeed)
		if err != nil {
			// Silently skip days with no data (expected for weekends/vacations)
			continue
		}

		if len(dayStats.ProjectBreakdown) == 0 {
			continue // Skip days with no activity
		}

		daysWithActivity++

		dateStr := currentDate.Format("2006-01-02")
		fmt.Printf("📅 %s - %d projects\n", dateStr, len(dayStats.ProjectBreakdown))

		// Convert per-project stats to API submissions for this day
		for _, proj := range dayStats.ProjectBreakdown {
			// Extract project name from path
			projectName := extractProjectName(proj.ProjectPath)

			// Calculate per-project tool breakdown (approximate based on command ratio)
			cmdRatio := float64(proj.Commands) / float64(dayStats.TotalCommands)

			// Get tool stats proportionally
			fileReads := 0
			fileEdits := 0
			fileWrites := 0
			searches := 0
			timeSavedReads := 0.0
			timeSavedEdits := 0.0
			timeSavedWrites := 0.0
			timeSavedSearches := 0.0

			for _, tool := range dayStats.ToolBreakdown {
				switch tool.ToolName {
				case "read":
					fileReads = int(float64(tool.Count) * cmdRatio)
					timeSavedReads = tool.TimeSavedMins * cmdRatio / 60.0 // Convert to hours
				case "edit":
					fileEdits = int(float64(tool.Count) * cmdRatio)
					timeSavedEdits = tool.TimeSavedMins * cmdRatio / 60.0
				case "write":
					fileWrites = int(float64(tool.Count) * cmdRatio)
					timeSavedWrites = tool.TimeSavedMins * cmdRatio / 60.0
				case "grep", "glob":
					searches += int(float64(tool.Count) * cmdRatio)
					timeSavedSearches += tool.TimeSavedMins * cmdRatio / 60.0
				}
			}

			timeSavedBash := proj.TimeSavedMins / 60.0 // Convert to hours
			timeSavedTotal := timeSavedBash + timeSavedReads + timeSavedEdits + timeSavedWrites + timeSavedSearches

			submission := ActivitySubmission{
				ID:                 fmt.Sprintf("%s-%s-%s", dateStr, devName, projectName),
				Date:               dateStr,
				Developer:          devName,
				ProjectPath:        proj.ProjectPath,
				ProjectName:        &projectName,
				BashCommands:       proj.Commands,
				FileReads:          fileReads,
				FileEdits:          fileEdits,
				FileWrites:         fileWrites,
				Searches:           searches,
				TimeSavedBash:      timeSavedBash,
				TimeSavedReads:     timeSavedReads,
				TimeSavedEdits:     timeSavedEdits,
				TimeSavedWrites:    timeSavedWrites,
				TimeSavedSearches:  timeSavedSearches,
				TimeSavedTotal:     timeSavedTotal,
				AvgComplexityScore: proj.AvgComplexityScore,
				SessionCount:       dayStats.SessionCount, // Total sessions, not per-project
				CharsTyped:         proj.Chars,
			}

			allSubmissions = append(allSubmissions, submission)

			// Print summary
			fmt.Printf("  📁 %s - Commands: %d | Time saved: %.1fh\n",
				projectName, submission.BashCommands, submission.TimeSavedTotal)
		}
	}

	if len(allSubmissions) == 0 {
		fmt.Println("\n📭 No OpenCode activity found in this date range.")
		return
	}

	fmt.Printf("\n📈 Summary: %d days with activity, %d total activities\n", daysWithActivity, len(allSubmissions))

	if *dryRun {
		fmt.Println("\n🔍 Dry run - would submit:")
		for _, sub := range allSubmissions {
			jsonData, _ := json.MarshalIndent(sub, "  ", "  ")
			fmt.Println(string(jsonData))
		}
		return
	}

	// Submit to API
	fmt.Printf("\n📤 Submitting %d activities to %s...\n", len(allSubmissions), baseURL)

	successCount := 0
	for _, sub := range allSubmissions {
		if err := submitActivity(baseURL, &sub); err != nil {
			fmt.Printf("  ❌ Failed to submit %s on %s: %v\n", *sub.ProjectName, sub.Date, err)
		} else {
			successCount++
		}
	}

	fmt.Printf("\n✨ Successfully submitted %d/%d activities\n", successCount, len(allSubmissions))
}

func findEarliestOpenCodeData() (time.Time, error) {
	dbPath, err := getOpenCodeDBPath()
	if err != nil {
		return time.Time{}, err
	}

	db, err := sql.Open("sqlite3", fmt.Sprintf("file:%s?mode=ro", dbPath))
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	var earliestMillis int64
	query := `
		SELECT MIN(p.time_created)
		FROM part p
		WHERE json_extract(p.data, '$.tool') = 'bash'
		  AND json_extract(p.data, '$.state.status') = 'completed'
		  AND json_extract(p.data, '$.state.input.command') IS NOT NULL
	`
	err = db.QueryRow(query).Scan(&earliestMillis)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to query earliest timestamp: %w", err)
	}

	if earliestMillis == 0 {
		return time.Time{}, fmt.Errorf("no data found in database")
	}

	return time.UnixMilli(earliestMillis), nil
}

func extractProjectName(path string) string {
	home := os.Getenv("HOME")

	// If path is exactly home directory, return "home"
	if path == home {
		return "home"
	}

	// Try to find git repository root
	gitRoot := findGitRoot(path)
	if gitRoot != "" {
		// Extract just the repository name
		repoName := gitRoot
		if strings.HasPrefix(repoName, home+"/") {
			repoName = strings.TrimPrefix(repoName, home+"/")
		}
		// Get the last directory of the repo path (the repo name itself)
		parts := strings.Split(strings.TrimSuffix(repoName, "/"), "/")
		if len(parts) > 0 {
			return parts[len(parts)-1]
		}
	}

	// Fallback: remove home directory prefix and use first directory
	if strings.HasPrefix(path, home+"/") {
		path = strings.TrimPrefix(path, home+"/")
	}

	// Handle edge cases
	if path == "." || path == "/" || path == "" {
		return "home"
	}

	// Split path into parts and use first directory
	parts := strings.Split(path, "/")
	if len(parts) == 0 {
		return "home"
	}

	// If path starts with "projects/", use the next directory as project name
	// e.g., "projects/main-project" -> "main-project"
	if len(parts) >= 2 && parts[0] == "projects" {
		return parts[1]
	}

	// Otherwise, use the first directory
	return parts[0]
}

func findGitRoot(path string) string {
	// Walk up the directory tree looking for .git
	current := path
	for {
		gitPath := current + "/.git"
		if _, err := os.Stat(gitPath); err == nil {
			return current
		}

		// Move to parent directory
		parent := current
		if idx := strings.LastIndex(current, "/"); idx > 0 {
			parent = current[:idx]
		} else {
			break
		}

		// Stop if we've reached the root or haven't moved
		if parent == current || parent == "" || parent == "/" {
			break
		}
		current = parent
	}
	return ""
}

func submitActivity(baseURL string, activity *ActivitySubmission) error {
	url := baseURL + "/api/v1/opencode/activity"

	jsonData, err := json.Marshal(activity)
	if err != nil {
		return fmt.Errorf("failed to marshal activity: %w", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errorMsg map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errorMsg)
		return fmt.Errorf("API returned %d: %v", resp.StatusCode, errorMsg)
	}

	return nil
}
