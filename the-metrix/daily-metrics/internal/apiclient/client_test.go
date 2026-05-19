package apiclient

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCheckHealth(t *testing.T) {
	tests := []struct {
		name        string
		statusCode  int
		wantHealthy bool
	}{
		{"healthy API", http.StatusOK, true},
		{"unhealthy API", http.StatusServiceUnavailable, false},
		{"not found", http.StatusNotFound, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/health" {
					w.WriteHeader(tt.statusCode)
				}
			}))
			defer server.Close()

			client, err := NewClient(server.URL, false)
			if err != nil {
				t.Fatalf("NewClient() error = %v", err)
			}
			gotHealthy := client.CheckHealth()

			if gotHealthy != tt.wantHealthy {
				t.Errorf("CheckHealth() = %v, want %v", gotHealthy, tt.wantHealthy)
			}
		})
	}
}

func TestGetMetrics(t *testing.T) {
	responseJSON := `{
		"date": "2026-04-17",
		"jira_issues": [{"type": "Task", "key": "PROJ-123", "summary": "Test", "status": "Done"}],
		"github_prs": [],
		"git_commits": 5,
		"lines_added": 100,
		"lines_deleted": 50
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/metrics" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(responseJSON))
		}
	}))
	defer server.Close()

	client, err := NewClient(server.URL, false)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	metrics, err := client.GetMetrics("2026-04-17", false)
	if err != nil {
		t.Fatalf("GetMetrics() error = %v", err)
	}

	if metrics.Date != "2026-04-17" {
		t.Errorf("Date = %v, want 2026-04-17", metrics.Date)
	}
	if metrics.GitCommits != 5 {
		t.Errorf("GitCommits = %v, want 5", metrics.GitCommits)
	}
	if len(metrics.JiraIssues) != 1 {
		t.Errorf("len(JiraIssues) = %v, want 1", len(metrics.JiraIssues))
	}
}

func TestNewClient_URLValidation(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{"valid http URL", "http://localhost:8080", false},
		{"valid https URL", "https://api.example.com", false},
		{"valid private IP", "http://192.168.1.1", false}, // No SSRF protection for internal tool
		{"invalid scheme", "ftp://localhost:8080", true},
		{"invalid URL", "not a url", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewClient(tt.url, false)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetMetrics_DateValidation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"date":"2026-04-17","jira_issues":[],"github_prs":[],"git_commits":0,"lines_added":0,"lines_deleted":0}`))
	}))
	defer server.Close()

	client, _ := NewClient(server.URL, false)

	tests := []struct {
		name    string
		date    string
		wantErr bool
	}{
		{"valid date", "2026-04-17", false},
		{"invalid format", "2026/04/17", true},
		{"too short", "2026-04", true},
		{"too long", "2026-04-17-extra", true},
		{"invalid month", "2026-13-01", true},
		{"invalid day", "2026-04-32", true},
		{"malformed", "aaaa-bb-cc", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := client.GetMetrics(tt.date, false)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetMetrics() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
