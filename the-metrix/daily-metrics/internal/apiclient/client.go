package apiclient

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type Client struct {
	BaseURL    string
	HTTPClient *http.Client
	Verbose    bool
}

type MetricsResponse struct {
	Date         string      `json:"date"`
	JiraIssues   []JiraIssue `json:"jira_issues"`
	GitHubPRs    []GitHubPR  `json:"github_prs"`
	GitCommits   int         `json:"git_commits"`
	LinesAdded   int         `json:"lines_added"`
	LinesDeleted int         `json:"lines_deleted"`
}

type JiraIssue struct {
	Type    string `json:"type"`
	Key     string `json:"key"`
	Summary string `json:"summary"`
	Status  string `json:"status"`
}

type GitHubPR struct {
	Number     int    `json:"number"`
	Title      string `json:"title"`
	State      string `json:"state"`
	UpdatedAt  string `json:"updated_at"`
	Repository struct {
		Name          string `json:"name"`
		NameWithOwner string `json:"name_with_owner"`
	} `json:"repository"`
}

func NewClient(baseURL string, verbose bool) (*Client, error) {
	// Validate URL to prevent SSRF
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	// Only allow http/https schemes
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return nil, fmt.Errorf("invalid URL scheme: %s (only http/https allowed)", parsedURL.Scheme)
	}

	// Note: No SSRF protection implemented - this is an internal development tool
	// where the API URL is explicitly provided by the developer via CLI flag

	return &Client{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second, // Longer timeout for metrics queries
		},
		Verbose: verbose,
	}, nil
}

func (c *Client) CheckHealth() bool {
	// Build URL properly
	healthURL, err := url.JoinPath(c.BaseURL, "/health")
	if err != nil {
		if c.Verbose {
			fmt.Printf("⚠️  Failed to build health URL: %v\n", err)
		}
		return false
	}

	if c.Verbose {
		fmt.Printf("🔍 Checking API health: %s\n", healthURL)
	}

	// Use shorter timeout for health checks
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(healthURL)
	if err != nil {
		if c.Verbose {
			fmt.Printf("⚠️  API health check failed: %v\n", err)
		}
		return false
	}
	defer resp.Body.Close()

	healthy := resp.StatusCode == http.StatusOK
	if c.Verbose {
		if healthy {
			fmt.Println("✅ API is healthy")
		} else {
			fmt.Printf("⚠️  API returned status: %d\n", resp.StatusCode)
		}
	}
	return healthy
}

func (c *Client) GetMetrics(date string, allRepos bool) (*MetricsResponse, error) {
	// Validate date format using time.Parse
	if _, err := time.Parse("2006-01-02", date); err != nil {
		return nil, fmt.Errorf("invalid date format: %s (expected YYYY-MM-DD)", date)
	}

	// Build URL properly with encoded query parameters
	metricsURL, err := url.JoinPath(c.BaseURL, "/api/metrics")
	if err != nil {
		return nil, fmt.Errorf("failed to build metrics URL: %w", err)
	}

	// Add query parameters using url.Values for proper encoding
	parsedURL, err := url.Parse(metricsURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse metrics URL: %w", err)
	}

	query := parsedURL.Query()
	query.Set("date", date)
	query.Set("all_repos", fmt.Sprintf("%t", allRepos))
	parsedURL.RawQuery = query.Encode()

	if c.Verbose {
		fmt.Printf("📡 API Request: %s\n", parsedURL.String())
	}

	resp, err := c.HTTPClient.Get(parsedURL.String())
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()
	defer io.Copy(io.Discard, resp.Body) // Drain remaining bytes to reuse connection

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var metrics MetricsResponse
	if err := json.NewDecoder(resp.Body).Decode(&metrics); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if c.Verbose {
		fmt.Println("✅ API response received")
	}

	return &metrics, nil
}
