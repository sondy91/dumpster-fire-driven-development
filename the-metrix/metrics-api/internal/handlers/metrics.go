package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/themetrix/metrics-api/pkg/calendar"
	"github.com/themetrix/metrics-api/pkg/gitops"
)

type GitMetricsResponse struct {
	StartDate   string `json:"start_date"`
	EndDate     string `json:"end_date"`
	Commits     int    `json:"commits"`
	Additions   int    `json:"additions"`
	Deletions   int    `json:"deletions"`
	NetChange   int    `json:"net_change"`
	WorkingDays int    `json:"working_days,omitempty"`
}

type GitHubMetricsResponse struct {
	StartDate string            `json:"start_date"`
	EndDate   string            `json:"end_date"`
	TotalPRs  int               `json:"total_prs"`
	MergedPRs int               `json:"merged_prs"`
	OpenPRs   int               `json:"open_prs"`
	ClosedPRs int               `json:"closed_prs"`
	PRs       []GitHubPRSummary `json:"prs,omitempty"`
}

type GitHubPRSummary struct {
	Number    int            `json:"number"`
	Title     string         `json:"title"`
	State     string         `json:"state"`
	CreatedAt string         `json:"created_at"`
	MergedAt  string         `json:"merged_at,omitempty"`
	Labels    []GitHubLabel  `json:"labels,omitempty"`
}

type GitHubLabel struct {
	Name string `json:"name"`
}

type JiraMetricsResponse struct {
	StartDate   string             `json:"start_date"`
	EndDate     string             `json:"end_date"`
	TotalIssues int                `json:"total_issues"`
	DoneIssues  int                `json:"done_issues"`
	Issues      []JiraIssueSummary `json:"issues,omitempty"`
}

type JiraIssueSummary struct {
	Key       string `json:"key"`
	Summary   string `json:"summary"`
	Status    string `json:"status"`
	IssueType string `json:"issue_type"`
	UpdatedAt string `json:"updated_at"`
}

func (s *Server) GetGitMetrics(w http.ResponseWriter, r *http.Request) {
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")
	eNumber := r.URL.Query().Get("e_number")

	if startDate == "" || endDate == "" {
		respondError(w, http.StatusBadRequest, "validation_error", "start_date and end_date are required (YYYY-MM-DD)")
		return
	}

	startTime, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		respondError(w, http.StatusBadRequest, "validation_error", "invalid start_date format (use YYYY-MM-DD)")
		return
	}

	endTime, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		respondError(w, http.StatusBadRequest, "validation_error", "invalid end_date format (use YYYY-MM-DD)")
		return
	}

	if startTime.After(endTime) {
		respondError(w, http.StatusBadRequest, "validation_error", "start_date must be before or equal to end_date")
		return
	}

	if endTime.After(time.Now()) {
		respondError(w, http.StatusBadRequest, "validation_error", "end_date cannot be in the future")
		return
	}

	cacheKey := fmt.Sprintf("git_metrics:%s:%s:%s", eNumber, startDate, endDate)

	if s.Config.Cache.Enabled {
		if cachedData, found, err := s.Storage.GetCache(cacheKey); err == nil && found {
			var response GitMetricsResponse
			if err := json.Unmarshal([]byte(cachedData), &response); err == nil {
				respondJSON(w, http.StatusOK, response)
				return
			}
		}
	}

	author := gitops.GetGitAuthor()
	if eNumber != "" {
		user, err := s.Storage.GetUser(eNumber)
		if err == nil && user != nil && user.Email != nil {
			author = *user.Email
		}
	}

	var repoPaths []string

	repos, err := s.Storage.ListRepos()
	if err == nil {
		for _, repo := range repos {
			if repo.RepoType == "work" {
				repoPaths = append(repoPaths, repo.Path)
			}
		}
	}

	if len(repoPaths) == 0 && s.Config != nil {
		projectsDir := filepath.Join(os.Getenv("HOME"), "projects")
		for _, repoName := range s.Config.Repos.Work {
			if strings.Contains(repoName, "..") || strings.Contains(repoName, "/") || strings.Contains(repoName, "\\") {
				log.Printf("WARN: Skipping invalid repo name with path traversal characters: %s", repoName)
				continue
			}
			repoPath := filepath.Join(projectsDir, repoName)
			if _, err := os.Stat(repoPath); err == nil {
				repoPaths = append(repoPaths, repoPath)
			}
		}
	}

	additions, deletions, err := gitops.CalculateLOCFromRepos(startDate, endDate, repoPaths, author)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "git_error", err.Error())
		return
	}

	commits := countCommits(startDate, endDate, repoPaths, author)

	calCfg := &calendar.Config{
		ScheduleBanchor: s.Config.ScheduleBanchor,
		Holidays2025:    s.Config.Holidays2025,
		Holidays2026:    s.Config.Holidays2026,
	}
	workingDays := calendar.CalculateWorkingDays(startDate, endDate, calCfg)

	response := GitMetricsResponse{
		StartDate:   startDate,
		EndDate:     endDate,
		Commits:     commits,
		Additions:   additions,
		Deletions:   deletions,
		NetChange:   additions - deletions,
		WorkingDays: workingDays,
	}

	if s.Config.Cache.Enabled {
		if responseData, err := json.Marshal(response); err == nil {
			var eNum *string
			if eNumber != "" {
				eNum = &eNumber
			}
			ttl := time.Duration(s.Config.Cache.TTLSeconds) * time.Second
			if err := s.Storage.SetCache("git_metrics", cacheKey, string(responseData), eNum, &startTime, &endTime, ttl); err != nil {
				log.Printf("WARN: Failed to cache git metrics for key %s: %v", cacheKey, err)
			}
		}
	}

	respondJSON(w, http.StatusOK, response)
}

func (s *Server) GetGitHubMetrics(w http.ResponseWriter, r *http.Request) {
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")

	if startDate == "" || endDate == "" {
		respondError(w, http.StatusBadRequest, "validation_error", "start_date and end_date are required (YYYY-MM-DD)")
		return
	}

	startTime, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		respondError(w, http.StatusBadRequest, "validation_error", "invalid start_date format (use YYYY-MM-DD)")
		return
	}

	endTime, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		respondError(w, http.StatusBadRequest, "validation_error", "invalid end_date format (use YYYY-MM-DD)")
		return
	}

	if startTime.After(endTime) {
		respondError(w, http.StatusBadRequest, "validation_error", "start_date must be before or equal to end_date")
		return
	}

	if endTime.After(time.Now()) {
		respondError(w, http.StatusBadRequest, "validation_error", "end_date cannot be in the future")
		return
	}

	cacheKey := fmt.Sprintf("github_metrics:%s:%s", startDate, endDate)

	if s.Config.Cache.Enabled {
		if cachedData, found, err := s.Storage.GetCache(cacheKey); err == nil && found {
			var response GitHubMetricsResponse
			if err := json.Unmarshal([]byte(cachedData), &response); err == nil {
				respondJSON(w, http.StatusOK, response)
				return
			}
		}
	}

	cmd := exec.Command("gh", "search", "prs",
		"--author", "@me",
		fmt.Sprintf("--created=%s..%s", startDate, endDate),
		"--json", "number,title,state,createdAt,closedAt,labels",
		"--limit", "1000")

	output, err := cmd.Output()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "github_error", "failed to query GitHub API")
		return
	}

	var prs []struct {
		Number    int    `json:"number"`
		Title     string `json:"title"`
		State     string `json:"state"`
		CreatedAt string `json:"createdAt"`
		ClosedAt  string `json:"closedAt"`
		Labels    []struct {
			Name string `json:"name"`
		} `json:"labels"`
	}

	if err := json.Unmarshal(output, &prs); err != nil {
		respondError(w, http.StatusInternalServerError, "parse_error", "failed to parse GitHub response")
		return
	}

	totalPRs := len(prs)
	mergedPRs := 0
	openPRs := 0
	closedPRs := 0

	prSummaries := make([]GitHubPRSummary, 0, len(prs))
	for _, pr := range prs {
		state := strings.ToUpper(pr.State)
		if state == "MERGED" {
			mergedPRs++
		} else if state == "OPEN" {
			openPRs++
		} else if state == "CLOSED" {
			closedPRs++
		}

		labels := make([]GitHubLabel, len(pr.Labels))
		for i, label := range pr.Labels {
			labels[i] = GitHubLabel{Name: label.Name}
		}

		prSummaries = append(prSummaries, GitHubPRSummary{
			Number:    pr.Number,
			Title:     pr.Title,
			State:     strings.ToLower(pr.State),
			CreatedAt: pr.CreatedAt,
			MergedAt:  pr.ClosedAt,
			Labels:    labels,
		})
	}

	response := GitHubMetricsResponse{
		StartDate: startDate,
		EndDate:   endDate,
		TotalPRs:  totalPRs,
		MergedPRs: mergedPRs,
		OpenPRs:   openPRs,
		ClosedPRs: closedPRs,
		PRs:       prSummaries,
	}

	if s.Config.Cache.Enabled {
		if responseData, err := json.Marshal(response); err == nil {
			ttl := time.Duration(s.Config.Cache.TTLSeconds) * time.Second
			if err := s.Storage.SetCache("github_metrics", cacheKey, string(responseData), nil, &startTime, &endTime, ttl); err != nil {
				log.Printf("WARN: Failed to cache GitHub metrics for key %s: %v", cacheKey, err)
			}
		}
	}

	respondJSON(w, http.StatusOK, response)
}

func (s *Server) GetJiraMetrics(w http.ResponseWriter, r *http.Request) {
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")

	if startDate == "" || endDate == "" {
		respondError(w, http.StatusBadRequest, "validation_error", "start_date and end_date are required (YYYY-MM-DD)")
		return
	}

	startTime, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		respondError(w, http.StatusBadRequest, "validation_error", "invalid start_date format (use YYYY-MM-DD)")
		return
	}

	endTime, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		respondError(w, http.StatusBadRequest, "validation_error", "invalid end_date format (use YYYY-MM-DD)")
		return
	}

	if startTime.After(endTime) {
		respondError(w, http.StatusBadRequest, "validation_error", "start_date must be before or equal to end_date")
		return
	}

	if endTime.After(time.Now()) {
		respondError(w, http.StatusBadRequest, "validation_error", "end_date cannot be in the future")
		return
	}

	cacheKey := fmt.Sprintf("jira_metrics:%s:%s", startDate, endDate)

	if s.Config.Cache.Enabled {
		if cachedData, found, err := s.Storage.GetCache(cacheKey); err == nil && found {
			var response JiraMetricsResponse
			if err := json.Unmarshal([]byte(cachedData), &response); err == nil {
				respondJSON(w, http.StatusOK, response)
				return
			}
		}
	}

	jql := fmt.Sprintf("assignee = currentUser() AND updated >= '%s' AND updated <= '%s'", startDate, endDate)

	cmd := exec.Command("jira", "issue", "list",
		"--jql", jql,
		"--raw")

	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			stderr := string(exitErr.Stderr)
			if strings.Contains(stderr, "No result found") {
				output = []byte("[]")
			} else {
				respondError(w, http.StatusInternalServerError, "jira_error", "failed to query Jira API")
				return
			}
		} else {
			respondError(w, http.StatusInternalServerError, "jira_error", "failed to query Jira API")
			return
		}
	}

	var jiraIssues []struct {
		Key    string `json:"key"`
		Fields struct {
			Summary   string `json:"summary"`
			IssueType struct {
				Name string `json:"name"`
			} `json:"issueType"`
			Status struct {
				Name string `json:"name"`
			} `json:"status"`
			Updated string `json:"updated"`
		} `json:"fields"`
	}

	if err := json.Unmarshal(output, &jiraIssues); err != nil {
		respondError(w, http.StatusInternalServerError, "parse_error", "failed to parse Jira response")
		return
	}

	totalIssues := len(jiraIssues)
	doneIssues := 0
	issues := make([]JiraIssueSummary, 0, len(jiraIssues))

	for _, issue := range jiraIssues {
		if strings.ToLower(issue.Fields.Status.Name) == "done" {
			doneIssues++
		}

		issues = append(issues, JiraIssueSummary{
			Key:       issue.Key,
			Summary:   issue.Fields.Summary,
			Status:    issue.Fields.Status.Name,
			IssueType: issue.Fields.IssueType.Name,
			UpdatedAt: issue.Fields.Updated,
		})
	}

	response := JiraMetricsResponse{
		StartDate:   startDate,
		EndDate:     endDate,
		TotalIssues: totalIssues,
		DoneIssues:  doneIssues,
		Issues:      issues,
	}

	if s.Config.Cache.Enabled {
		if responseData, err := json.Marshal(response); err == nil {
			ttl := time.Duration(s.Config.Cache.TTLSeconds) * time.Second
			if err := s.Storage.SetCache("jira_metrics", cacheKey, string(responseData), nil, &startTime, &endTime, ttl); err != nil {
				log.Printf("WARN: Failed to cache Jira metrics for key %s: %v", cacheKey, err)
			}
		}
	}

	respondJSON(w, http.StatusOK, response)
}

func countCommits(startDate, endDate string, repoPaths []string, author string) int {
	totalCommits := 0

	for _, repoPath := range repoPaths {
		cmd := exec.Command("git", "log",
			fmt.Sprintf("--author=%s", author),
			fmt.Sprintf("--since=%s 00:00", startDate),
			fmt.Sprintf("--until=%s 23:59", endDate),
			"--oneline")
		cmd.Dir = repoPath

		output, err := cmd.Output()
		if err != nil {
			continue
		}

		lines := strings.Split(strings.TrimSpace(string(output)), "\n")
		if len(lines) > 0 && lines[0] != "" {
			totalCommits += len(lines)
		}
	}

	return totalCommits
}
