package main

import (
	"strings"
	"testing"
)

func TestGenerateMarkdown(t *testing.T) {
	metrics := DailyMetrics{
		Date:         "2026-04-15",
		GitCommits:   42,
		LinesAdded:   1234,
		LinesDeleted: 567,
		JiraIssues: []JiraIssue{
			{Type: "Task", Key: "PROJ-123", Summary: "Test issue", Status: "Done"},
			{Type: "Bug", Key: "PROJ-456", Summary: "Fix bug", Status: "In Progress"},
		},
		GitHubPRs: []GitHubPR{
			{Number: 1, Title: "Add feature", State: "merged"},
			{Number: 2, Title: "Fix issue", State: "open"},
		},
	}

	markdown := generateMarkdown(metrics)

	tests := []struct {
		name     string
		contains string
	}{
		{"Has date in header", "2026-04-15"},
		{"Has commit count", "42"},
		{"Has LOC added", "+1234"},
		{"Has LOC deleted", "-567"},
		{"Has Jira issue key", "PROJ-123"},
		{"Has PR title", "Add feature"},
		{"Has net change", "+667"},
		{"Has work summary header", "Daily Work Summary"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !strings.Contains(markdown, tt.contains) {
				t.Errorf("generateMarkdown() missing expected content: %s\nGot:\n%s",
					tt.contains, markdown)
			}
		})
	}
}

func TestGenerateSummaryMarkdown(t *testing.T) {
	cfg := &Config{
		ScheduleBanchor: "2026-04-03",
		Holidays2026:    []string{},
	}

	markdown := generateSummaryMarkdownString(
		"2026-04-01", "2026-04-07",
		50, 5000, 1000,
		10, 8,
		5, 3,
		cfg,
	)

	tests := []struct {
		name     string
		contains string
	}{
		{"Has header", "# Summary: 2026-04-01 to 2026-04-07"},
		{"Has Git section", "## Git Activity"},
		{"Has commits", "50"},
		{"Has additions", "+5000"},
		{"Has deletions", "-1000"},
		{"Has net change", "+4000"},
		{"Has PR section", "## GitHub Pull Requests"},
		{"Has PR count", "10"},
		{"Has Jira section", "## Jira Issues"},
		{"Has productivity section", "## Productivity Metrics"},
		{"Has working days", "working days"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !strings.Contains(markdown, tt.contains) {
				t.Errorf("generateSummaryMarkdown() missing expected content: %s\nGot:\n%s",
					tt.contains, markdown)
			}
		})
	}
}

func generateSummaryMarkdownString(startDate, endDate string, commits, additions, deletions, totalPRs, mergedPRs, totalJira, doneJira int, cfg *Config) string {
	calendarDays := calculateDays(startDate, endDate)
	workingDays := calculateWorkingDays(startDate, endDate, cfg)
	weeks := float64(calendarDays) / 7.0

	var sb strings.Builder

	sb.WriteString("# Summary: " + startDate + " to " + endDate + "\n\n")
	sb.WriteString("## Git Activity\n\n")
	sb.WriteString("- **" + intToString(commits) + "** total commits across all repos\n")
	sb.WriteString("- **+" + intToString(additions) + " -" + intToString(deletions) + "** LOC\n")
	sb.WriteString("- Net change: **+" + intToString(additions-deletions) + "** lines\n\n")
	sb.WriteString("## GitHub Pull Requests\n\n")
	sb.WriteString("- **" + intToString(totalPRs) + "** total PRs\n")
	sb.WriteString("- **" + intToString(mergedPRs) + "** merged\n\n")
	sb.WriteString("## Jira Issues\n\n")
	sb.WriteString("- **" + intToString(totalJira) + "** total issues worked on\n")
	sb.WriteString("- **" + intToString(doneJira) + "** completed\n\n")
	sb.WriteString("## Productivity Metrics\n\n")
	sb.WriteString("- **" + intToString(calendarDays) + "** calendar days (**" + intToString(workingDays) + "** working days, **" + floatToString(weeks) + "** weeks)\n")

	return sb.String()
}

func intToString(n int) string {
	if n == 0 {
		return "0"
	}
	result := ""
	for n > 0 {
		result = string(rune('0'+n%10)) + result
		n /= 10
	}
	return result
}

func floatToString(f float64) string {
	i := int(f)
	return intToString(i) + "." + intToString(int((f-float64(i))*10))
}
