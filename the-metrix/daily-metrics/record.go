package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func handleRecord() {
	// Create flagset for record subcommand
	recordFlags := flag.NewFlagSet("record", flag.ExitOnError)
	note := recordFlags.String("note", "", "Note/context to record")
	jiraID := recordFlags.String("jira", "", "Jira issue ID (e.g., PROJ-123)")
	prNum := recordFlags.String("pr", "", "PR number")
	noTimestamp := recordFlags.Bool("no-timestamp", false, "Omit timestamp from note")

	recordFlags.Parse(os.Args[2:])

	if *note == "" {
		fmt.Println("❌ Note is required. Usage: daily-metrics record --note \"Your note here\" [--jira PROJ-123] [--pr 42]")
		os.Exit(1)
	}

	// Get today's date
	date := time.Now().Format("2006-01-02")

	// Build the note entry
	var entry strings.Builder

	// Add timestamp if not disabled
	if !*noTimestamp {
		timestamp := time.Now().Format("3:04 PM")
		entry.WriteString(fmt.Sprintf("**%s** - ", timestamp))
	}

	// Add main note
	entry.WriteString(*note)

	// Add Jira link if provided
	if *jiraID != "" {
		jiraBaseURL := os.Getenv("JIRA_BASE_URL")
		if jiraBaseURL == "" {
			jiraBaseURL = "https://your-company.atlassian.net"
		}
		entry.WriteString(fmt.Sprintf(" ([%s](%s/browse/%s))", *jiraID, jiraBaseURL, *jiraID))
	}

	// Add PR reference if provided
	if *prNum != "" {
		entry.WriteString(fmt.Sprintf(" (PR #%s)", *prNum))
	}

	entry.WriteString("\n")

	// Get daily note path
	dailyNotePath := getDailyNotePath(date)

	// Append to daily note
	if err := appendToDailyNote(dailyNotePath, entry.String()); err != nil {
		fmt.Printf("❌ Failed to append to daily note: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Recorded to daily note: %s\n", dailyNotePath)
}

func getDailyNotePath(date string) string {
	vaultPath := os.ExpandEnv("$HOME/projects/SecondBrain")
	return filepath.Join(vaultPath, "2. Area (Long-term responsibilities)", "Daily Notes", date+".md")
}

func appendToDailyNote(path, content string) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Check if file exists
	_, err := os.Stat(path)
	fileExists := !os.IsNotExist(err)

	// If file doesn't exist, create with header
	if !fileExists {
		header := fmt.Sprintf("# Daily Notes - %s\n\n## 🤖 Session Notes\n\n", filepath.Base(path[:len(path)-3]))
		if err := os.WriteFile(path, []byte(header), 0644); err != nil {
			return err
		}
	}

	// Read existing content
	existingContent, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	existingStr := string(existingContent)

	// Check if Session Notes section exists
	if !strings.Contains(existingStr, "## 🤖 Session Notes") {
		// Add section header at the end
		existingStr += "\n## 🤖 Session Notes\n\n"
	}

	// Find the Session Notes section and append
	sessionIdx := strings.Index(existingStr, "## 🤖 Session Notes")
	if sessionIdx == -1 {
		// This shouldn't happen, but handle gracefully
		existingStr += "\n## 🤖 Session Notes\n\n"
		sessionIdx = strings.Index(existingStr, "## 🤖 Session Notes")
	}

	// Find next section or end of file
	afterSession := existingStr[sessionIdx+len("## 🤖 Session Notes"):]
	nextSectionIdx := strings.Index(afterSession, "\n## ")

	if nextSectionIdx == -1 {
		// No next section, append at end
		existingStr += content
	} else {
		// Insert before next section
		insertPos := sessionIdx + len("## 🤖 Session Notes") + nextSectionIdx
		existingStr = existingStr[:insertPos] + content + existingStr[insertPos:]
	}

	// Write back
	return os.WriteFile(path, []byte(existingStr), 0644)
}
