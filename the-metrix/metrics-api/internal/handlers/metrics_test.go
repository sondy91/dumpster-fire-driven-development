package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetGitMetricsValidation(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	tests := []struct {
		name       string
		startDate  string
		endDate    string
		wantStatus int
	}{
		{
			name:       "missing start_date",
			startDate:  "",
			endDate:    "2026-04-16",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "missing end_date",
			startDate:  "2026-04-01",
			endDate:    "",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid start_date format",
			startDate:  "2026/04/01",
			endDate:    "2026-04-16",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid end_date format",
			startDate:  "2026-04-01",
			endDate:    "04-16-2026",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/api/v1/metrics/git?start_date=" + tt.startDate + "&end_date=" + tt.endDate
			req := httptest.NewRequest("GET", url, nil)
			rr := httptest.NewRecorder()

			server.GetGitMetrics(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d: %s", tt.wantStatus, rr.Code, rr.Body.String())
			}
		})
	}
}

func TestGetGitMetricsSuccess(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/api/v1/metrics/git?start_date=2026-04-01&end_date=2026-04-16", nil)
	rr := httptest.NewRecorder()

	server.GetGitMetrics(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d: %s", http.StatusOK, rr.Code, rr.Body.String())
	}

	var response GitMetricsResponse
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.StartDate != "2026-04-01" {
		t.Errorf("expected start_date 2026-04-01, got %s", response.StartDate)
	}

	if response.EndDate != "2026-04-16" {
		t.Errorf("expected end_date 2026-04-16, got %s", response.EndDate)
	}

	if response.WorkingDays <= 0 {
		t.Errorf("expected working_days > 0, got %d", response.WorkingDays)
	}
}

func TestGetGitHubMetricsValidation(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/api/v1/metrics/github?start_date=&end_date=", nil)
	rr := httptest.NewRecorder()

	server.GetGitHubMetrics(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestGetJiraMetricsValidation(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/api/v1/metrics/jira?start_date=&end_date=", nil)
	rr := httptest.NewRecorder()

	server.GetJiraMetrics(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}
