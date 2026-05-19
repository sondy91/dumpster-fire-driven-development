package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gorilla/mux"
)

func TestCreateRepo(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	tmpDir := t.TempDir()
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatalf("failed to create test git dir: %v", err)
	}

	reqBody := CreateRepoRequest{
		Name:     "test-repo",
		Path:     tmpDir,
		RepoType: "work",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/repos", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	server.CreateRepo(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d: %s", http.StatusCreated, rr.Code, rr.Body.String())
	}

	var response RepoResponse
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Name != "test-repo" {
		t.Errorf("expected name test-repo, got %s", response.Name)
	}

	if response.RepoType != "work" {
		t.Errorf("expected repo_type work, got %s", response.RepoType)
	}
}

func TestCreateRepoValidation(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	tmpDir := t.TempDir()
	gitDir := filepath.Join(tmpDir, ".git")
	os.MkdirAll(gitDir, 0755)

	tests := []struct {
		name       string
		reqBody    CreateRepoRequest
		wantStatus int
	}{
		{
			name:       "missing name",
			reqBody:    CreateRepoRequest{Path: "/path", RepoType: "work"},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "missing path",
			reqBody:    CreateRepoRequest{Name: "repo", RepoType: "work"},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid repo_type",
			reqBody:    CreateRepoRequest{Name: "repo", Path: tmpDir, RepoType: "invalid"},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "defaults to work",
			reqBody:    CreateRepoRequest{Name: "repo", Path: tmpDir},
			wantStatus: http.StatusCreated,
		},
		{
			name:       "path does not exist",
			reqBody:    CreateRepoRequest{Name: "repo", Path: "/nonexistent/path"},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "path is not a git repo",
			reqBody:    CreateRepoRequest{Name: "repo", Path: t.TempDir()},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.reqBody)
			req := httptest.NewRequest("POST", "/api/v1/repos", bytes.NewBuffer(body))
			rr := httptest.NewRecorder()

			server.CreateRepo(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d: %s", tt.wantStatus, rr.Code, rr.Body.String())
			}
		})
	}
}

func TestGetRepo(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	tmpDir := t.TempDir()
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatalf("failed to create test git dir: %v", err)
	}

	reqBody := CreateRepoRequest{
		Name:     "test-repo",
		Path:     tmpDir,
		RepoType: "work",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/repos", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()
	server.CreateRepo(rr, req)

	req = httptest.NewRequest("GET", "/api/v1/repos/home/user/projects/test-repo", nil)
	rr = httptest.NewRecorder()
	req = mux.SetURLVars(req, map[string]string{"path": tmpDir})

	server.GetRepo(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var response RepoResponse
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Name != "test-repo" {
		t.Errorf("expected name test-repo, got %s", response.Name)
	}
}

func TestGetRepoNotFound(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/api/v1/repos/nonexistent", nil)
	rr := httptest.NewRecorder()
	req = mux.SetURLVars(req, map[string]string{"path": "/nonexistent/path"})

	server.GetRepo(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, rr.Code)
	}
}

func TestListRepos(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	tmpDir1 := t.TempDir()
	gitDir1 := filepath.Join(tmpDir1, ".git")
	if err := os.MkdirAll(gitDir1, 0755); err != nil {
		t.Fatalf("failed to create test git dir: %v", err)
	}

	tmpDir2 := t.TempDir()
	gitDir2 := filepath.Join(tmpDir2, ".git")
	if err := os.MkdirAll(gitDir2, 0755); err != nil {
		t.Fatalf("failed to create test git dir: %v", err)
	}

	repos := []CreateRepoRequest{
		{Name: "repo1", Path: tmpDir1, RepoType: "work"},
		{Name: "repo2", Path: tmpDir2, RepoType: "private"},
	}

	for _, repo := range repos {
		body, _ := json.Marshal(repo)
		req := httptest.NewRequest("POST", "/api/v1/repos", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()
		server.CreateRepo(rr, req)
	}

	req := httptest.NewRequest("GET", "/api/v1/repos", nil)
	rr := httptest.NewRecorder()

	server.ListRepos(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var response []RepoResponse
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(response) != 2 {
		t.Errorf("expected 2 repos, got %d", len(response))
	}
}
