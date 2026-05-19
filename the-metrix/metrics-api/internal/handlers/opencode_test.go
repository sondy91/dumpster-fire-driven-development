package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/themetrix/metrics-api/internal/storage"
)

func TestOpenCodeSubmitSession(t *testing.T) {
	// Create in-memory storage for testing
	store, err := storage.New(":memory:")
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer store.Close()

	handler := NewOpenCodeHandler(store)

	tests := []struct {
		name           string
		payload        OpenCodeSubmitRequest
		expectedStatus int
		expectError    bool
	}{
		{
			name: "valid session",
			payload: OpenCodeSubmitRequest{
				ID:                "test-session-1",
				SavedAt:           "2026-04-21T10:00:00Z",
				Developer:         "testuser",
				Project:           "test-project",
				Sprint:            "Sprint 12",
				Bash:              100,
				Reads:             50,
				Edits:             25,
				Searches:          10,
				TimeSavedBash:     200.0,
				TimeSavedReads:    25.0,
				TimeSavedEdits:    50.0,
				TimeSavedSearches: 10.0,
				TimeSavedTotal:    285.0,
				SessionDuration:   60,
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name: "missing developer",
			payload: OpenCodeSubmitRequest{
				ID:      "test-session-2",
				SavedAt: "2026-04-21T10:00:00Z",
				Project: "test-project",
				Sprint:  "Sprint 12",
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name: "missing project",
			payload: OpenCodeSubmitRequest{
				ID:        "test-session-3",
				SavedAt:   "2026-04-21T10:00:00Z",
				Developer: "testuser",
				Sprint:    "Sprint 12",
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name: "missing sprint",
			payload: OpenCodeSubmitRequest{
				ID:        "test-session-4",
				SavedAt:   "2026-04-21T10:00:00Z",
				Developer: "testuser",
				Project:   "test-project",
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name: "auto-generated ID",
			payload: OpenCodeSubmitRequest{
				SavedAt:        "2026-04-21T10:00:00Z",
				Developer:      "testuser",
				Project:        "test-project",
				Sprint:         "Sprint 12",
				Bash:           50,
				TimeSavedTotal: 100.0,
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/opencode/submit", bytes.NewBuffer(payload))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			handler.SubmitSession(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d. Body: %s", tt.expectedStatus, w.Code, w.Body.String())
			}

			if !tt.expectError {
				var resp map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
					t.Errorf("Failed to parse response: %v", err)
				}

				if success, ok := resp["success"].(bool); !ok || !success {
					t.Errorf("Expected success=true in response, got %v", resp)
				}

				if id, ok := resp["id"].(string); !ok || id == "" {
					t.Errorf("Expected non-empty ID in response, got %v", resp)
				}
			}
		})
	}
}

func TestOpenCodeGetSessions(t *testing.T) {
	// Create in-memory storage and populate with test data
	store, err := storage.New(":memory:")
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer store.Close()

	// Insert test sessions
	testSessions := []*storage.OpenCodeSession{
		{
			ID:                "session-1",
			SavedAt:           "2026-04-21T10:00:00Z",
			Developer:         "user1",
			Project:           "project-a",
			Sprint:            "Sprint 12",
			Bash:              100,
			Reads:             50,
			Edits:             25,
			Searches:          10,
			TimeSavedBash:     200.0,
			TimeSavedReads:    25.0,
			TimeSavedEdits:    50.0,
			TimeSavedSearches: 10.0,
			TimeSavedTotal:    285.0,
			SessionDuration:   60,
		},
		{
			ID:                "session-2",
			SavedAt:           "2026-04-21T11:00:00Z",
			Developer:         "user1",
			Project:           "project-b",
			Sprint:            "Sprint 12",
			Bash:              80,
			Reads:             40,
			Edits:             20,
			Searches:          5,
			TimeSavedBash:     160.0,
			TimeSavedReads:    20.0,
			TimeSavedEdits:    40.0,
			TimeSavedSearches: 5.0,
			TimeSavedTotal:    225.0,
			SessionDuration:   45,
		},
		{
			ID:                "session-3",
			SavedAt:           "2026-04-21T12:00:00Z",
			Developer:         "user2",
			Project:           "project-a",
			Sprint:            "Sprint 11",
			Bash:              120,
			Reads:             60,
			Edits:             30,
			Searches:          15,
			TimeSavedBash:     240.0,
			TimeSavedReads:    30.0,
			TimeSavedEdits:    60.0,
			TimeSavedSearches: 15.0,
			TimeSavedTotal:    345.0,
			SessionDuration:   75,
		},
	}

	for _, sess := range testSessions {
		if err := store.CreateOpenCodeSession(sess); err != nil {
			t.Fatalf("Failed to insert test session: %v", err)
		}
	}

	handler := NewOpenCodeHandler(store)

	tests := []struct {
		name          string
		queryParams   map[string]string
		expectedCount int
	}{
		{
			name:          "all sessions",
			queryParams:   map[string]string{},
			expectedCount: 3,
		},
		{
			name: "filter by developer",
			queryParams: map[string]string{
				"developer": "user1",
			},
			expectedCount: 2,
		},
		{
			name: "filter by project",
			queryParams: map[string]string{
				"project": "project-a",
			},
			expectedCount: 2,
		},
		{
			name: "filter by sprint",
			queryParams: map[string]string{
				"sprint": "Sprint 12",
			},
			expectedCount: 2,
		},
		{
			name: "filter by multiple params",
			queryParams: map[string]string{
				"developer": "user1",
				"sprint":    "Sprint 12",
			},
			expectedCount: 2,
		},
		{
			name: "limit results",
			queryParams: map[string]string{
				"limit": "1",
			},
			expectedCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/opencode/sessions", nil)
			q := req.URL.Query()
			for k, v := range tt.queryParams {
				q.Add(k, v)
			}
			req.URL.RawQuery = q.Encode()

			w := httptest.NewRecorder()
			handler.GetSessions(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
			}

			var resp map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Errorf("Failed to parse response: %v", err)
			}

			count, ok := resp["count"].(float64)
			if !ok {
				t.Errorf("Expected count in response, got %v", resp)
			}

			if int(count) != tt.expectedCount {
				t.Errorf("Expected %d sessions, got %d", tt.expectedCount, int(count))
			}

			sessions, ok := resp["sessions"].([]interface{})
			if !ok {
				t.Errorf("Expected sessions array in response, got %v", resp)
			}

			if len(sessions) != tt.expectedCount {
				t.Errorf("Expected %d sessions in array, got %d", tt.expectedCount, len(sessions))
			}
		})
	}
}

func TestOpenCodeMethodNotAllowed(t *testing.T) {
	store, err := storage.New(":memory:")
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer store.Close()

	handler := NewOpenCodeHandler(store)

	tests := []struct {
		name    string
		method  string
		handler func(http.ResponseWriter, *http.Request)
	}{
		{"GET on SubmitSession", http.MethodGet, handler.SubmitSession},
		{"PUT on SubmitSession", http.MethodPut, handler.SubmitSession},
		{"POST on GetSessions", http.MethodPost, handler.GetSessions},
		{"DELETE on GetSessions", http.MethodDelete, handler.GetSessions},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/test", nil)
			w := httptest.NewRecorder()

			tt.handler(w, req)

			if w.Code != http.StatusMethodNotAllowed {
				t.Errorf("Expected status 405, got %d", w.Code)
			}
		})
	}
}

func TestOpenCodeInvalidJSON(t *testing.T) {
	store, err := storage.New(":memory:")
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer store.Close()

	handler := NewOpenCodeHandler(store)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/opencode/submit", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.SubmitSession(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 for invalid JSON, got %d", w.Code)
	}
}
