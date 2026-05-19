package main

import (
	"fmt"
	"os"
	"strings"
)

func generateMarkdown(m DailyMetrics) string {
	var sb strings.Builder
	
	jiraBaseURL := os.Getenv("JIRA_BASE_URL")
	if jiraBaseURL == "" {
		jiraBaseURL = "https://your-company.atlassian.net"
	}

	sb.WriteString(fmt.Sprintf("# Daily Work Summary - %s\n\n", m.Date))

	// Summary stats at top
	sb.WriteString("## 📊 At a Glance\n\n")
	sb.WriteString(fmt.Sprintf("- **%d** commits across work repos\n", m.GitCommits))
	sb.WriteString(fmt.Sprintf("- **+%d -%d** lines of code\n", m.LinesAdded, m.LinesDeleted))

	prMerged := 0
	prOpen := 0
	for _, pr := range m.GitHubPRs {
		if pr.State == "merged" {
			prMerged++
		} else if pr.State == "open" {
			prOpen++
		}
	}
	sb.WriteString(fmt.Sprintf("- **%d** PRs merged, **%d** created/updated\n", prMerged, len(m.GitHubPRs)))

	jiraDone := 0
	jiraInReview := 0
	jiraInProgress := 0
	for _, issue := range m.JiraIssues {
		switch issue.Status {
		case "Done":
			jiraDone++
		case "Review", "In Review":
			jiraInReview++
		case "Progress", "In Progress":
			jiraInProgress++
		}
	}
	sb.WriteString(fmt.Sprintf("- **%d** Jira issues completed, **%d** in review, **%d** in progress\n", jiraDone, jiraInReview, jiraInProgress))

	sb.WriteString("\n---\n\n")

	// Jira Issues Section
	if len(m.JiraIssues) > 0 {
		sb.WriteString("## ✅ Jira Activity\n\n")

		// Group by status
		statusGroups := make(map[string][]JiraIssue)
		for _, issue := range m.JiraIssues {
			statusGroups[issue.Status] = append(statusGroups[issue.Status], issue)
		}

		// Done items first
		if issues, ok := statusGroups["Done"]; ok {
			sb.WriteString("### Completed\n\n")
			for _, issue := range issues {
				sb.WriteString(fmt.Sprintf("- **[%s](%s/browse/%s)** - %s\n", issue.Key, jiraBaseURL, issue.Key, issue.Summary))
			}
			sb.WriteString("\n")
		}

		// In Review
		if issues, ok := statusGroups["Review"]; ok {
			sb.WriteString("### In Review\n\n")
			for _, issue := range issues {
				sb.WriteString(fmt.Sprintf("- **[%s](%s/browse/%s)** - %s\n", issue.Key, jiraBaseURL, issue.Key, issue.Summary))
			}
			sb.WriteString("\n")
		}
		if issues, ok := statusGroups["In Review"]; ok {
			for _, issue := range issues {
				sb.WriteString(fmt.Sprintf("- **[%s](%s/browse/%s)** - %s\n", issue.Key, jiraBaseURL, issue.Key, issue.Summary))
			}
		}

		// In Progress
		if issues, ok := statusGroups["Progress"]; ok {
			sb.WriteString("### In Progress\n\n")
			for _, issue := range issues {
				sb.WriteString(fmt.Sprintf("- **[%s](%s/browse/%s)** - %s\n", issue.Key, jiraBaseURL, issue.Key, issue.Summary))
			}
			sb.WriteString("\n")
		}
		if issues, ok := statusGroups["In Progress"]; ok {
			for _, issue := range issues {
				sb.WriteString(fmt.Sprintf("- **[%s](%s/browse/%s)** - %s\n", issue.Key, jiraBaseURL, issue.Key, issue.Summary))
			}
		}

		// Other statuses
		for status, issues := range statusGroups {
			if status != "Done" && status != "Review" && status != "In Review" && status != "Progress" && status != "In Progress" {
				sb.WriteString(fmt.Sprintf("### %s\n\n", status))
				for _, issue := range issues {
					sb.WriteString(fmt.Sprintf("- **[%s](%s/browse/%s)** - %s\n", issue.Key, jiraBaseURL, issue.Key, issue.Summary))
				}
				sb.WriteString("\n")
			}
		}
	}

	// GitHub PRs Section
	if len(m.GitHubPRs) > 0 {
		sb.WriteString("## 🔀 Pull Requests\n\n")

		// Merged PRs
		var merged []GitHubPR
		var open []GitHubPR
		var closed []GitHubPR

		for _, pr := range m.GitHubPRs {
			switch pr.State {
			case "merged":
				merged = append(merged, pr)
			case "open":
				open = append(open, pr)
			case "closed":
				closed = append(closed, pr)
			}
		}

		if len(merged) > 0 {
			sb.WriteString("### ✓ Merged\n\n")
			for _, pr := range merged {
				repoName := pr.Repository.Name
				sb.WriteString(fmt.Sprintf("- **#%d**: %s\n", pr.Number, pr.Title))
				sb.WriteString(fmt.Sprintf("  - Repo: `%s`\n", repoName))
			}
			sb.WriteString("\n")
		}

		if len(open) > 0 {
			sb.WriteString("### ○ Open/Updated\n\n")
			for _, pr := range open {
				repoName := pr.Repository.Name
				sb.WriteString(fmt.Sprintf("- **#%d**: %s\n", pr.Number, pr.Title))
				sb.WriteString(fmt.Sprintf("  - Repo: `%s`\n", repoName))
			}
			sb.WriteString("\n")
		}

		if len(closed) > 0 {
			sb.WriteString("### ✗ Closed\n\n")
			for _, pr := range closed {
				repoName := pr.Repository.Name
				sb.WriteString(fmt.Sprintf("- **#%d**: %s\n", pr.Number, pr.Title))
				sb.WriteString(fmt.Sprintf("  - Repo: `%s`\n", repoName))
			}
			sb.WriteString("\n")
		}
	}

	// Git commits summary
	if m.GitCommits > 0 {
		sb.WriteString("## 💾 Code Contributions\n\n")
		sb.WriteString(fmt.Sprintf("- **%d** commits today\n", m.GitCommits))
		sb.WriteString(fmt.Sprintf("- **+%d** lines added\n", m.LinesAdded))
		sb.WriteString(fmt.Sprintf("- **-%d** lines removed\n", m.LinesDeleted))
		sb.WriteString(fmt.Sprintf("- **Net change**: %+d lines\n", m.LinesAdded-m.LinesDeleted))
		sb.WriteString("\n")
	}

	return sb.String()
}
