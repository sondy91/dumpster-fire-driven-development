package main

import (
"database/sql"
"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	_ "github.com/mattn/go-sqlite3"
)

var ErrNoData = errors.New("no OpenCode commands found in date range")

type BashCommand struct {
	SessionID          string
	Command            string
	Description        string
	TimeCreated        int64
	CharCount          int
	Directory          string
	ComplexityMultiplier float64
}

type ProjectStats struct {
	ProjectPath          string  `json:"project_path"`
	Commands             int     `json:"commands"`
	Chars                int     `json:"chars"`
	TimeSavedMins        float64 `json:"time_saved_minutes"`
	AvgComplexityScore   float64 `json:"avg_complexity_score"`
}

type EditOperation struct {
	FilePath    string
	OldLength   int
	NewLength   int
	NetChange   int
	TimeCreated int64
}

type ToolStats struct {
	ToolName      string  `json:"tool_name"`
	Count         int     `json:"count"`
	Chars         int     `json:"chars"`
	TimeSavedMins float64 `json:"time_saved_minutes"`
}

type OpenCodeStats struct {
	Commands              []BashCommand
	TotalCommands         int         `json:"total_commands"`
	TotalChars            int         `json:"total_chars"`
	TimeSavedMins         float64     `json:"command_time_saved_minutes"`
	SessionCount          int         `json:"session_count"`
	StartDate             string      `json:"start_date"`
	EndDate               string      `json:"end_date"`
	ProjectBreakdown      []ProjectStats `json:"projects"`
	Edits                 []EditOperation
	TotalEdits            int     `json:"total_edits"`
	TotalCharsEdited      int     `json:"total_chars_edited"`
	EditTimeSavedMin      float64     `json:"edit_time_saved_minutes"`
	AvgComplexityScore    float64     `json:"avg_complexity_score"`
	WeightedTimeSavedMins float64     `json:"weighted_time_saved_minutes"`
	TotalCharsWritten     int         `json:"total_chars_written"`
	WriteTimeSavedMin     float64     `json:"write_time_saved_minutes"`
	ToolBreakdown         []ToolStats `json:"tool_breakdown"`
}

func getOpenCodeDBPath() (string, error) {
	home := os.Getenv("HOME")
	if home == "" {
		return "", fmt.Errorf("HOME environment variable not set")
	}

	dbPath := filepath.Join(home, ".local", "share", "opencode", "opencode.db")

	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return "", fmt.Errorf("OpenCode database not found at %s. Have you used OpenCode yet?", dbPath)
	}

	return dbPath, nil
}

func queryBashCommands(since, until time.Time) ([]BashCommand, error) {
	dbPath, err := getOpenCodeDBPath()
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite3", fmt.Sprintf("file:%s?mode=ro", dbPath))
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	sinceMillis := since.UnixMilli()
	untilMillis := until.UnixMilli()

	query := `
		SELECT
			p.session_id,
			json_extract(p.data, '$.state.input.command') as command,
			json_extract(p.data, '$.state.input.description') as description,
			p.time_created,
			COALESCE(s.directory, '') as directory
		FROM part p
		LEFT JOIN session s ON p.session_id = s.id
		WHERE json_extract(p.data, '$.tool') = 'bash'
		  AND json_extract(p.data, '$.state.status') = 'completed'
		  AND p.time_created >= ?
		  AND p.time_created <= ?
		  AND command IS NOT NULL
		ORDER BY p.time_created ASC
	`

	rows, err := db.Query(query, sinceMillis, untilMillis)
	if err != nil {
		return nil, fmt.Errorf("failed to query commands: %w", err)
	}
	defer rows.Close()

	var commands []BashCommand
	for rows.Next() {
		var cmd BashCommand
		var description sql.NullString

		if err := rows.Scan(&cmd.SessionID, &cmd.Command, &description, &cmd.TimeCreated, &cmd.Directory); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		if description.Valid {
			cmd.Description = description.String
		}

		cmd.CharCount = len(cmd.Command)
		commands = append(commands, cmd)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return commands, nil
}

func queryEditOperations(since, until time.Time) ([]EditOperation, error) {
	dbPath, err := getOpenCodeDBPath()
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite3", fmt.Sprintf("file:%s?mode=ro", dbPath))
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	sinceMillis := since.UnixMilli()
	untilMillis := until.UnixMilli()

	query := `
		SELECT
			json_extract(data, '$.state.input.filePath') as filePath,
			LENGTH(json_extract(data, '$.state.input.oldString')) as oldLength,
			LENGTH(json_extract(data, '$.state.input.newString')) as newLength,
			time_created
		FROM part
		WHERE json_extract(data, '$.tool') = 'edit'
		  AND json_extract(data, '$.state.status') = 'completed'
		  AND time_created >= ?
		  AND time_created <= ?
		  AND filePath IS NOT NULL
		ORDER BY time_created ASC
	`

	rows, err := db.Query(query, sinceMillis, untilMillis)
	if err != nil {
		return nil, fmt.Errorf("failed to query edits: %w", err)
	}
	defer rows.Close()

	var edits []EditOperation
	for rows.Next() {
		var edit EditOperation

		if err := rows.Scan(&edit.FilePath, &edit.OldLength, &edit.NewLength, &edit.TimeCreated); err != nil {
			return nil, fmt.Errorf("failed to scan edit row: %w", err)
		}

		edit.NetChange = edit.NewLength - edit.OldLength
		edits = append(edits, edit)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating edit rows: %w", err)
	}

	return edits, nil
}

func queryWriteOperations(since, until time.Time) (int, error) {
	dbPath, err := getOpenCodeDBPath()
	if err != nil {
		return 0, err
	}

	db, err := sql.Open("sqlite3", fmt.Sprintf("file:%s?mode=ro", dbPath))
	if err != nil {
		return 0, fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	sinceMillis := since.UnixMilli()
	untilMillis := until.UnixMilli()

	query := `
		SELECT
			COALESCE(SUM(LENGTH(json_extract(data, '$.state.input.content'))), 0) as totalChars
		FROM part
		WHERE json_extract(data, '$.tool') = 'write'
		  AND json_extract(data, '$.state.status') = 'completed'
		  AND time_created >= ?
		  AND time_created <= ?
	`

	var totalChars int
	err = db.QueryRow(query, sinceMillis, untilMillis).Scan(&totalChars)
	if err != nil {
		return 0, fmt.Errorf("failed to query write operations: %w", err)
	}

	return totalChars, nil
}

// queryToolUsage aggregates typing overhead metrics for grep/glob/read/write/edit/task tools.
// The "Chars" metric represents typing overhead saved (characters to invoke the tool), NOT work performed.
// For each tool, chars = length of the filepath/pattern you would have typed to invoke the tool:
//   - grep: pattern length + optional path length
//   - glob: pattern length
//   - read: filepath length
//   - write: filepath length
//   - edit: filepath length + 10 (estimated overhead for specifying oldString/newString parameters
//           beyond just the filepath; based on typical edit command structure)
//   - task: 0 (task typing overhead not easily quantified by filepath/pattern length)
// This design measures typing overhead saved to invoke the tool, not the content generated
// or processed by the tool (which is tracked separately via queryWriteOperations for write tool).
func queryToolUsage(since, until time.Time) (map[string]ToolStats, error) {
	dbPath, err := getOpenCodeDBPath()
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite3", fmt.Sprintf("file:%s?mode=ro", dbPath))
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	sinceMillis := since.UnixMilli()
	untilMillis := until.UnixMilli()

	query := `
		SELECT
			json_extract(data, '$.tool') as tool,
			COUNT(*) as count,
			SUM(CASE
				WHEN json_extract(data, '$.tool') = 'grep' THEN COALESCE(LENGTH(json_extract(data, '$.state.input.pattern')), 0) + COALESCE(LENGTH(json_extract(data, '$.state.input.path')), 0)
				WHEN json_extract(data, '$.tool') = 'glob' THEN COALESCE(LENGTH(json_extract(data, '$.state.input.pattern')), 0)
				WHEN json_extract(data, '$.tool') = 'read' THEN COALESCE(LENGTH(json_extract(data, '$.state.input.filePath')), 0)
				WHEN json_extract(data, '$.tool') = 'write' THEN COALESCE(LENGTH(json_extract(data, '$.state.input.filePath')), 0)
				WHEN json_extract(data, '$.tool') = 'edit' THEN COALESCE(LENGTH(json_extract(data, '$.state.input.filePath')), 0) + 10
				ELSE 0
			END) as totalChars
		FROM part
		WHERE json_extract(data, '$.state.status') = 'completed'
		  AND time_created >= ?
		  AND time_created <= ?
		  AND tool IS NOT NULL
		  AND json_extract(data, '$.tool') IN ('grep', 'glob', 'read', 'write', 'edit', 'task')
		GROUP BY tool
	`

	rows, err := db.Query(query, sinceMillis, untilMillis)
	if err != nil {
		return nil, fmt.Errorf("failed to query tool usage: %w", err)
	}
	defer rows.Close()

	toolStats := make(map[string]ToolStats)
	for rows.Next() {
		var tool string
		var count, chars int

		if err := rows.Scan(&tool, &count, &chars); err != nil {
			return nil, fmt.Errorf("failed to scan tool usage row: %w", err)
		}

		toolStats[tool] = ToolStats{
			ToolName: tool,
			Count:    count,
			Chars:    chars,
		}
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating tool usage rows: %w", err)
	}

	return toolStats, nil
}

func calculateTimeSaved(totalChars int, wpm int) float64 {
	if wpm <= 0 {
		wpm = 50
	}

	charsPerMinute := float64(wpm * 5)
	minutesSaved := float64(totalChars) / charsPerMinute

	return minutesSaved
}

const (
	// Complexity scoring constants
	complexityBaseMultiplier = 1.0
	complexityLengthDivisor  = 10.0

	// Pattern bonus multipliers
	pipeBonus        = 0.5
	conditionalBonus = 0.5
	heredocBonus     = 0.5
	toolBonus        = 0.3  // jq, awk, sed

	// Display formatting
	maxProjectPathLength = 45
	projectPathTruncate  = 42
)

var (
	// Compile regexes once for performance
	conditionalPattern = regexp.MustCompile(`\b(if|then|else|elif|case|esac|while|until|for|do|done)\b`)
	heredocPattern     = regexp.MustCompile(`<<-?\s*['"]?\w+['"]?`)
	jqPattern          = regexp.MustCompile(`\bjq\b`)
	awkPattern         = regexp.MustCompile(`\bawk\b`)
	sedPattern         = regexp.MustCompile(`\bsed\b`)
)

func calculateComplexityMultiplier(command string) float64 {
	cmdLen := len(command)
	if cmdLen == 0 {
	return complexityBaseMultiplier
	}

	lengthFactor := 1.0 + math.Log10(float64(cmdLen)/complexityLengthDivisor)
	if lengthFactor < 1.0 {
		lengthFactor = 1.0
	}

	patternBonus := 0.0

	if strings.Contains(command, "|") {
		patternBonus += pipeBonus
	}

	if conditionalPattern.MatchString(command) {
		patternBonus += conditionalBonus
	}

	if heredocPattern.MatchString(command) {
		patternBonus += heredocBonus
	}

	if jqPattern.MatchString(command) {
		patternBonus += toolBonus
	}
	if awkPattern.MatchString(command) {
		patternBonus += toolBonus
	}
	if sedPattern.MatchString(command) {
		patternBonus += toolBonus
	}


	return complexityBaseMultiplier * lengthFactor * (1.0 + patternBonus)
}

func analyzeOpenCodeUsage(since, until time.Time, wpm int) (*OpenCodeStats, error) {
	commands, err := queryBashCommands(since, until)
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}

	if len(commands) == 0 {
		return nil, ErrNoData
	}

	totalChars := 0
	totalWeightedChars := 0.0
	sessionMap := make(map[string]bool)
	projectMap := make(map[string]*ProjectStats)
	projectComplexityMap := make(map[string][]float64)

	for i := range commands {
		multiplier := calculateComplexityMultiplier(commands[i].Command)
		commands[i].ComplexityMultiplier = multiplier

		totalChars += commands[i].CharCount
		totalWeightedChars += float64(commands[i].CharCount) * multiplier
		sessionMap[commands[i].SessionID] = true

		projectPath := commands[i].Directory
		if projectPath == "" {
			projectPath = "(unknown)"
		}

		if _, exists := projectMap[projectPath]; !exists {
			projectMap[projectPath] = &ProjectStats{
				ProjectPath: projectPath,
			}
		}
		projectMap[projectPath].Commands++
		projectMap[projectPath].Chars += commands[i].CharCount
		projectComplexityMap[projectPath] = append(projectComplexityMap[projectPath], multiplier)
	}

	for path, proj := range projectMap {
		proj.TimeSavedMins = calculateTimeSaved(proj.Chars, wpm)

		scores := projectComplexityMap[path]
		if len(scores) > 0 {
			sum := 0.0
			for _, s := range scores {
				sum += s
			}
			proj.AvgComplexityScore = sum / float64(len(scores))
		}
	}

	projectBreakdown := make([]ProjectStats, 0, len(projectMap))
	for _, proj := range projectMap {
		projectBreakdown = append(projectBreakdown, *proj)
	}

	sort.Slice(projectBreakdown, func(i, j int) bool {
		return projectBreakdown[i].TimeSavedMins > projectBreakdown[j].TimeSavedMins
	})

	timeSaved := calculateTimeSaved(totalChars, wpm)

	edits, err := queryEditOperations(since, until)
	if err != nil {
		return nil, err
	}

	// Count AI-generated characters from edit operations
	// Always count NewLength - that's the content AI actually wrote
	totalCharsEdited := 0
	for _, edit := range edits {
		totalCharsEdited += edit.NewLength
	}

	editTimeSaved := calculateTimeSaved(totalCharsEdited, wpm)

	totalCharsWritten, err := queryWriteOperations(since, until)
	if err != nil {
		return nil, fmt.Errorf("failed to query write operations: %w", err)
	}

	writeTimeSaved := calculateTimeSaved(totalCharsWritten, wpm)

	toolStatsMap, err := queryToolUsage(since, until)
	if err != nil {
		return nil, fmt.Errorf("failed to query tool usage: %w", err)
	}

	toolBreakdown := make([]ToolStats, 0, len(toolStatsMap))
	for _, stats := range toolStatsMap {
		stats.TimeSavedMins = calculateTimeSaved(stats.Chars, wpm)
		toolBreakdown = append(toolBreakdown, stats)
	}

	sort.Slice(toolBreakdown, func(i, j int) bool {
		return toolBreakdown[i].TimeSavedMins > toolBreakdown[j].TimeSavedMins
	})

	avgComplexity := 0.0
	if len(commands) > 0 {
		totalMultiplier := 0.0
		for _, cmd := range commands {
			totalMultiplier += cmd.ComplexityMultiplier
		}
		avgComplexity = totalMultiplier / float64(len(commands))
	}

	weightedTimeSaved := calculateTimeSaved(int(totalWeightedChars), wpm)

	stats := &OpenCodeStats{
		Commands:              commands,
		TotalCommands:         len(commands),
		TotalChars:            totalChars,
		TimeSavedMins:         timeSaved,
		SessionCount:          len(sessionMap),
		StartDate:             since.Format("Jan 2"),
		EndDate:               until.Format("Jan 2, 2006"),
		ProjectBreakdown:      projectBreakdown,
		Edits:                 edits,
		TotalEdits:            len(edits),
		TotalCharsEdited:      totalCharsEdited,
		EditTimeSavedMin:      editTimeSaved,
		AvgComplexityScore:    avgComplexity,
		WeightedTimeSavedMins: weightedTimeSaved,
		TotalCharsWritten:     totalCharsWritten,
		WriteTimeSavedMin:     writeTimeSaved,
		ToolBreakdown:         toolBreakdown,
	}

	return stats, nil
}

func formatOpenCodeStats(stats *OpenCodeStats, wpm int, since, until time.Time) {
	fmt.Println("╔══════════════════════════════════════════════════════════╗")
	title := "OpenCode AI Time Savings"
	if stats.StartDate == stats.EndDate {
		dateStr := stats.EndDate
		fullTitle := fmt.Sprintf("%s - %s", title, dateStr)
		padding := 58 - len(fullTitle)
		fmt.Printf("║ %s%s║\n", fullTitle, strings.Repeat(" ", padding))
	} else {
		dateRange := fmt.Sprintf("%s - %s", stats.StartDate, stats.EndDate)
		fullTitle := fmt.Sprintf("%s - %s", title, dateRange)
		padding := 58 - len(fullTitle)
		fmt.Printf("║ %s%s║\n", fullTitle, strings.Repeat(" ", padding))
	}
	fmt.Println("╚══════════════════════════════════════════════════════════╝")
	fmt.Println()

	fmt.Println("📊 Activity Summary")
	fmt.Println("─────────────────────────────────────────────────────────────")
	fmt.Printf("Sessions analyzed:      %6d\n", stats.SessionCount)
	fmt.Printf("Bash commands:          %6d (%s chars)\n", stats.TotalCommands, formatNumber(stats.TotalChars))
	fmt.Printf("File edits:             %6d (%s chars generated)\n", stats.TotalEdits, formatNumber(stats.TotalCharsEdited))
	fmt.Printf("File writes:            %6s chars generated\n", formatNumber(stats.TotalCharsWritten))
	fmt.Println()

	fmt.Println("⏱️  Time Savings")
	fmt.Println("─────────────────────────────────────────────────────────────")
	fmt.Printf("Your typing speed:       %d WPM (%d chars/min)\n", wpm, wpm*5)
	fmt.Printf("Avg complexity score:    %.2fx\n", stats.AvgComplexityScore)
	fmt.Printf("Command typing saved:    %s\n", formatDuration(stats.TimeSavedMins))
	fmt.Printf("Weighted command time:   %s\n", formatDuration(stats.WeightedTimeSavedMins))
	fmt.Printf("Code editing saved:      %s\n", formatDuration(stats.EditTimeSavedMin))
	fmt.Printf("Code writing saved:      %s\n", formatDuration(stats.WriteTimeSavedMin))
	totalSaved := stats.TimeSavedMins + stats.EditTimeSavedMin + stats.WriteTimeSavedMin
	fmt.Printf("Total time saved:        %s\n", formatDuration(totalSaved))

	if stats.TotalCommands > 0 || stats.TotalEdits > 0 {
		days := int(until.Sub(since).Hours()/24) + 1
		if days < 1 {
			days = 1
		}

		avgPerDay := totalSaved / float64(days)
		fmt.Printf("Daily average:           %s/day\n", formatDuration(avgPerDay))
	}

	fmt.Println()

	if len(stats.ProjectBreakdown) > 1 {
		fmt.Println("📁 Project Breakdown")
		fmt.Println("─────────────────────────────────────────────────────────────")

		displayCount := len(stats.ProjectBreakdown)
		// Show all projects (removed 10 project limit)

		for i := 0; i < displayCount; i++ {
			proj := stats.ProjectBreakdown[i]
			projectName := proj.ProjectPath

			home := os.Getenv("HOME")
			if strings.HasPrefix(projectName, home) {
				projectName = "~" + strings.TrimPrefix(projectName, home)
			}

			// Truncate long paths (handles multi-byte UTF-8 correctly)
			if utf8.RuneCountInString(projectName) > maxProjectPathLength {
				runes := []rune(projectName)
				projectName = "..." + string(runes[len(runes)-projectPathTruncate:])
			}

			fmt.Printf("%-*s %s (%.2fx, %d cmds)\n",
				maxProjectPathLength,
				projectName,
				formatDuration(proj.TimeSavedMins),
				proj.AvgComplexityScore,
				proj.Commands)
		}

		fmt.Println()
	}

	if len(stats.ToolBreakdown) > 0 {
		fmt.Println("🔧 Tool Usage Breakdown (typing overhead saved)")
		fmt.Println("─────────────────────────────────────────────────────────────")

		for _, tool := range stats.ToolBreakdown {
			if tool.Count > 0 {
				fmt.Printf("%-15s %6d calls, %6s chars, %s\n",
					tool.ToolName+":",
					tool.Count,
					formatNumber(tool.Chars),
					formatDuration(tool.TimeSavedMins))
			}
		}
		fmt.Println()
	}

	totalWork := stats.TimeSavedMins + stats.EditTimeSavedMin + stats.WriteTimeSavedMin
	fmt.Printf("💡 During this period, AI handled ~%s of work\n", formatDuration(totalWork))
}

func formatDuration(minutes float64) string {
	if minutes < 1.0 {
		return fmt.Sprintf("%.0f seconds", minutes*60)
	} else if minutes < 60.0 {
		return fmt.Sprintf("%.1f minutes", minutes)
	} else {
		hours := minutes / 60.0
		return fmt.Sprintf("%.1f hours", hours)
	}
}

func formatNumber(n int) string {
	s := fmt.Sprintf("%d", n)
	if len(s) <= 3 {
		return s
	}

	var result []byte
	for i, digit := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			result = append(result, ',')
		}
		result = append(result, byte(digit))
	}
	return string(result)
}

func getOpenCodeSummaryStats(startDate, endDate string) (commands, edits int, timeSaved float64, err error) {
	since, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return 0, 0, 0, err
	}
	until, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return 0, 0, 0, err
	}
	// Set to end of day in local timezone
	until = time.Date(until.Year(), until.Month(), until.Day(), 23, 59, 59, 999999999, time.Local)

	cfg, err := loadConfig()
	if err != nil {
		return 0, 0, 0, err
	}

	wpm := cfg.TypingSpeedWPM
	if wpm == 0 {
		wpm = 50
	}

	stats, err := analyzeOpenCodeUsage(since, until, wpm)
	if err != nil {
		return 0, 0, 0, err
	}

	totalTime := stats.TimeSavedMins + stats.EditTimeSavedMin
	return stats.TotalCommands, stats.TotalEdits, totalTime, nil
}

func querySessionsInRange(since, until time.Time) ([]string, error) {
	dbPath, err := getOpenCodeDBPath()
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite3", fmt.Sprintf("file:%s?mode=ro", dbPath))
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	sinceMillis := since.UnixMilli()
	untilMillis := until.UnixMilli()

	query := `
		SELECT DISTINCT session_id
		FROM part
		WHERE json_extract(data, '$.tool') = 'bash'
		  AND time_created >= ?
		  AND time_created <= ?
	`

	rows, err := db.Query(query, sinceMillis, untilMillis)
	if err != nil {
		return nil, fmt.Errorf("failed to query sessions: %w", err)
	}
	defer rows.Close()

	var sessions []string
	for rows.Next() {
		var sessionID string
		if err := rows.Scan(&sessionID); err != nil {
			return nil, fmt.Errorf("failed to scan session: %w", err)
		}
		sessions = append(sessions, sessionID)
	}

	return sessions, nil
}
func exportJSON(stats *OpenCodeStats, outputPath string) error {
	data, err := json.MarshalIndent(stats, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// Use 0600 permissions for privacy (ADR-001: personal metrics should remain private)
	// Owner-only read/write prevents accidental exposure on shared systems
	if err := os.WriteFile(outputPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write JSON file: %w", err)
	}

	return nil
}

func exportMarkdown(stats *OpenCodeStats, wpm int, outputPath string) error {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# OpenCode AI Time Savings\n\n"))
	sb.WriteString(fmt.Sprintf("**Date Range:** %s - %s\n\n", stats.StartDate, stats.EndDate))

	sb.WriteString("## Summary\n\n")
	sb.WriteString(fmt.Sprintf("- **Sessions analyzed:** %d\n", stats.SessionCount))
	sb.WriteString(fmt.Sprintf("- **Bash commands:** %d\n", stats.TotalCommands))
	sb.WriteString(fmt.Sprintf("- **Total characters:** %s\n", formatNumber(stats.TotalChars)))
	sb.WriteString(fmt.Sprintf("- **File edits:** %d\n", stats.TotalEdits))
	sb.WriteString(fmt.Sprintf("- **Characters edited:** %s\n\n", formatNumber(stats.TotalCharsEdited)))

	sb.WriteString("## Time Savings\n\n")
	sb.WriteString(fmt.Sprintf("- **Typing speed:** %d WPM (%d chars/min)\n", wpm, wpm*5))
	sb.WriteString(fmt.Sprintf("- **Command typing saved:** %s\n", formatDuration(stats.TimeSavedMins)))
	sb.WriteString(fmt.Sprintf("- **Code generation saved:** %s\n", formatDuration(stats.EditTimeSavedMin)))
	totalSaved := stats.TimeSavedMins + stats.EditTimeSavedMin
	sb.WriteString(fmt.Sprintf("- **Total time saved:** %s\n\n", formatDuration(totalSaved)))

	if len(stats.ProjectBreakdown) > 0 {
		sb.WriteString("## Project Breakdown\n\n")
		sb.WriteString("| Project | Time Saved | Commands |\n")
		sb.WriteString("|---------|------------|----------|\n")

		displayCount := len(stats.ProjectBreakdown)
		if displayCount > 10 {
			displayCount = 10
		}

		for i := 0; i < displayCount; i++ {
			proj := stats.ProjectBreakdown[i]
			projectName := proj.ProjectPath

			home := os.Getenv("HOME")
			if strings.HasPrefix(projectName, home) {
				projectName = "~" + strings.TrimPrefix(projectName, home)
			}

			sb.WriteString(fmt.Sprintf("| %s | %s | %d |\n",
				projectName,
				formatDuration(proj.TimeSavedMins),
				proj.Commands))
		}

		if len(stats.ProjectBreakdown) > 10 {
			sb.WriteString(fmt.Sprintf("\n*...and %d more projects*\n", len(stats.ProjectBreakdown)-10))
		}
	}

	// Use 0600 permissions for privacy (ADR-001: personal metrics should remain private)
	// Owner-only read/write prevents accidental exposure on shared systems
	if err := os.WriteFile(outputPath, []byte(sb.String()), 0600); err != nil {
		return fmt.Errorf("failed to write markdown file: %w", err)
	}

	return nil
}
