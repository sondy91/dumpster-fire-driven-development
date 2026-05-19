package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gorilla/mux"
	"github.com/themetrix/metrics-api/internal/storage"
	"github.com/themetrix/metrics-api/pkg/config"
)

func setupTestServer(t *testing.T) (*Server, func()) {
	t.Helper()
	dbPath := "./test_handlers.db"

	store, err := storage.New(dbPath)
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}

	cfg := &config.MetricsConfig{}
	server := NewServer(store, cfg)

	cleanup := func() {
		store.Close()
		os.Remove(dbPath)
	}

	return server, cleanup
}

func TestCreateUser(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	jiraID := "712020:abcd1234"
	githubUser := "testuser"
	name := "Test User"
	email := "test@example.com"

	reqBody := CreateUserRequest{
		ENumber:        "e12345678",
		JiraAccountID:  &jiraID,
		GitHubUsername: &githubUser,
		Name:           &name,
		Email:          &email,
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	server.CreateUser(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d: %s", http.StatusCreated, rr.Code, rr.Body.String())
	}

	var response UserResponse
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.ENumber != "e12345678" {
		t.Errorf("expected e_number e12345678, got %s", response.ENumber)
	}

	if response.GitHubUsername == nil || *response.GitHubUsername != githubUser {
		t.Errorf("expected github_username %s, got %v", githubUser, response.GitHubUsername)
	}
}

func TestGetUser(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	jiraID := "712020:abcd1234"
	user := &storage.User{
		ENumber:       "e12345678",
		JiraAccountID: &jiraID,
	}
	server.Storage.CreateUser(user)

	req := httptest.NewRequest("GET", "/api/v1/users/e12345678", nil)
	rr := httptest.NewRecorder()

	req = mux.SetURLVars(req, map[string]string{"e_number": "e12345678"})

	server.GetUser(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var response UserResponse
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.ENumber != "e12345678" {
		t.Errorf("expected e_number e12345678, got %s", response.ENumber)
	}
}

func TestGetUserNotFound(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/api/v1/users/e99999999", nil)
	rr := httptest.NewRecorder()

	req = mux.SetURLVars(req, map[string]string{"e_number": "e99999999"})

	server.GetUser(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, rr.Code)
	}
}

func TestListUsers(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	jiraID1 := "712020:abcd1234"
	jiraID2 := "712020:efgh5678"
	server.Storage.CreateUser(&storage.User{ENumber: "e11111111", JiraAccountID: &jiraID1})
	server.Storage.CreateUser(&storage.User{ENumber: "e22222222", JiraAccountID: &jiraID2})

	req := httptest.NewRequest("GET", "/api/v1/users", nil)
	rr := httptest.NewRecorder()

	server.ListUsers(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var response []UserResponse
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(response) != 2 {
		t.Errorf("expected 2 users, got %d", len(response))
	}
}

func TestUpdateUser(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	jiraID := "712020:abcd1234"
	githubUser := "olduser"
	server.Storage.CreateUser(&storage.User{
		ENumber:        "e12345678",
		JiraAccountID:  &jiraID,
		GitHubUsername: &githubUser,
	})

	newGithub := "newuser"
	reqBody := CreateUserRequest{
		ENumber:        "e12345678",
		JiraAccountID:  &jiraID,
		GitHubUsername: &newGithub,
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("PUT", "/api/v1/users/e12345678", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	req = mux.SetURLVars(req, map[string]string{"e_number": "e12345678"})

	server.UpdateUser(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var response UserResponse
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.GitHubUsername == nil || *response.GitHubUsername != newGithub {
		t.Errorf("expected github_username %s, got %v", newGithub, response.GitHubUsername)
	}
}

func TestDeleteUser(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	jiraID := "712020:abcd1234"
	server.Storage.CreateUser(&storage.User{
		ENumber:       "e12345678",
		JiraAccountID: &jiraID,
	})

	req := httptest.NewRequest("DELETE", "/api/v1/users/e12345678", nil)
	rr := httptest.NewRecorder()

	req = mux.SetURLVars(req, map[string]string{"e_number": "e12345678"})

	server.DeleteUser(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("expected status %d, got %d", http.StatusNoContent, rr.Code)
	}

	user, _ := server.Storage.GetUser("e12345678")
	if user != nil {
		t.Error("user should be deleted")
	}
}
