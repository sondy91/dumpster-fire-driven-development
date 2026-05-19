package storage

import (
	"os"
	"testing"
	"time"
)

func setupTestDB(t *testing.T) (*Storage, func()) {
	t.Helper()
	dbPath := "./test_metrics.db"

	storage, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}

	cleanup := func() {
		storage.Close()
		os.Remove(dbPath)
	}

	return storage, cleanup
}

func TestUserCRUD(t *testing.T) {
	storage, cleanup := setupTestDB(t)
	defer cleanup()

	eNumber := "e40057167"
	jiraID := "712020:abcd1234-5678-90ef-ghij-klmnopqrstuv"
	githubUser := "testuser"
	name := "Test User"
	email := "test@example.com"

	t.Run("CreateUser", func(t *testing.T) {
		user := &User{
			ENumber:        eNumber,
			JiraAccountID:  &jiraID,
			GitHubUsername: &githubUser,
			Name:           &name,
			Email:          &email,
		}

		err := storage.CreateUser(user)
		if err != nil {
			t.Fatalf("failed to create user: %v", err)
		}
	})

	t.Run("GetUser", func(t *testing.T) {
		user, err := storage.GetUser(eNumber)
		if err != nil {
			t.Fatalf("failed to get user: %v", err)
		}
		if user == nil {
			t.Fatal("user not found")
		}
		if user.ENumber != eNumber {
			t.Errorf("expected e_number %s, got %s", eNumber, user.ENumber)
		}
		if user.GitHubUsername == nil || *user.GitHubUsername != githubUser {
			t.Errorf("expected github_username %s, got %v", githubUser, user.GitHubUsername)
		}
	})

	t.Run("UpdateUser", func(t *testing.T) {
		newGithub := "newuser"
		user := &User{
			ENumber:        eNumber,
			JiraAccountID:  &jiraID,
			GitHubUsername: &newGithub,
			Name:           &name,
			Email:          &email,
		}

		err := storage.UpdateUser(user)
		if err != nil {
			t.Fatalf("failed to update user: %v", err)
		}

		updated, err := storage.GetUser(eNumber)
		if err != nil {
			t.Fatalf("failed to get updated user: %v", err)
		}
		if updated.GitHubUsername == nil || *updated.GitHubUsername != newGithub {
			t.Errorf("expected github_username %s, got %v", newGithub, updated.GitHubUsername)
		}
	})

	t.Run("ListUsers", func(t *testing.T) {
		users, err := storage.ListUsers()
		if err != nil {
			t.Fatalf("failed to list users: %v", err)
		}
		if len(users) != 1 {
			t.Errorf("expected 1 user, got %d", len(users))
		}
	})

	t.Run("DeleteUser", func(t *testing.T) {
		err := storage.DeleteUser(eNumber)
		if err != nil {
			t.Fatalf("failed to delete user: %v", err)
		}

		user, err := storage.GetUser(eNumber)
		if err != nil {
			t.Fatalf("failed to get user after delete: %v", err)
		}
		if user != nil {
			t.Error("user should be nil after delete")
		}
	})
}

func TestRepoCRUD(t *testing.T) {
	storage, cleanup := setupTestDB(t)
	defer cleanup()

	t.Run("CreateRepo", func(t *testing.T) {
		repo := &Repo{
			Name:     "test-repo",
			Path:     "/home/user/projects/test-repo",
			RepoType: "work",
		}

		err := storage.CreateRepo(repo)
		if err != nil {
			t.Fatalf("failed to create repo: %v", err)
		}
		if repo.ID == 0 {
			t.Error("expected repo ID to be set after create")
		}
	})

	t.Run("GetRepo", func(t *testing.T) {
		repo, err := storage.GetRepo("/home/user/projects/test-repo")
		if err != nil {
			t.Fatalf("failed to get repo: %v", err)
		}
		if repo == nil {
			t.Fatal("repo not found")
		}
		if repo.Name != "test-repo" {
			t.Errorf("expected name test-repo, got %s", repo.Name)
		}
		if repo.RepoType != "work" {
			t.Errorf("expected repo_type work, got %s", repo.RepoType)
		}
	})

	t.Run("ListRepos", func(t *testing.T) {
		repos, err := storage.ListRepos()
		if err != nil {
			t.Fatalf("failed to list repos: %v", err)
		}
		if len(repos) != 1 {
			t.Errorf("expected 1 repo, got %d", len(repos))
		}
	})
}

func TestConfigOperations(t *testing.T) {
	storage, cleanup := setupTestDB(t)
	defer cleanup()

	t.Run("SetAndGetConfig", func(t *testing.T) {
		err := storage.SetConfig("test_key", "test_value")
		if err != nil {
			t.Fatalf("failed to set config: %v", err)
		}

		value, err := storage.GetConfig("test_key")
		if err != nil {
			t.Fatalf("failed to get config: %v", err)
		}
		if value != "test_value" {
			t.Errorf("expected test_value, got %s", value)
		}
	})

	t.Run("UpdateConfig", func(t *testing.T) {
		err := storage.SetConfig("test_key", "updated_value")
		if err != nil {
			t.Fatalf("failed to update config: %v", err)
		}

		value, err := storage.GetConfig("test_key")
		if err != nil {
			t.Fatalf("failed to get updated config: %v", err)
		}
		if value != "updated_value" {
			t.Errorf("expected updated_value, got %s", value)
		}
	})

	t.Run("GetNonexistentConfig", func(t *testing.T) {
		value, err := storage.GetConfig("nonexistent_key")
		if err != nil {
			t.Fatalf("failed to get nonexistent config: %v", err)
		}
		if value != "" {
			t.Errorf("expected empty string for nonexistent key, got %s", value)
		}
	})
}

func TestCacheOperations(t *testing.T) {
	storage, cleanup := setupTestDB(t)
	defer cleanup()

	eNumber := "e40057167"
	startDate := time.Now().AddDate(0, 0, -7)
	endDate := time.Now()

	t.Run("SetAndGetCache", func(t *testing.T) {
		err := storage.SetCache("git_metrics", "test_key", `{"commits": 10}`, &eNumber, &startDate, &endDate, 5*time.Minute)
		if err != nil {
			t.Fatalf("failed to set cache: %v", err)
		}

		data, found, err := storage.GetCache("test_key")
		if err != nil {
			t.Fatalf("failed to get cache: %v", err)
		}
		if !found {
			t.Error("cache entry should be found")
		}
		if data != `{"commits": 10}` {
			t.Errorf("expected cached data, got %s", data)
		}
	})

	t.Run("GetExpiredCache", func(t *testing.T) {
		err := storage.SetCache("git_metrics", "expired_key", `{"commits": 5}`, &eNumber, &startDate, &endDate, 1*time.Millisecond)
		if err != nil {
			t.Fatalf("failed to set cache: %v", err)
		}

		time.Sleep(10 * time.Millisecond)

		data, found, err := storage.GetCache("expired_key")
		if err != nil {
			t.Fatalf("failed to get expired cache: %v", err)
		}
		if found {
			t.Error("expired cache entry should not be found")
		}
		if data != "" {
			t.Errorf("expected empty data for expired cache, got %s", data)
		}
	})

	t.Run("DeleteExpiredCache", func(t *testing.T) {
		err := storage.SetCache("git_metrics", "to_delete", `{"commits": 1}`, &eNumber, &startDate, &endDate, 1*time.Millisecond)
		if err != nil {
			t.Fatalf("failed to set cache: %v", err)
		}

		time.Sleep(10 * time.Millisecond)

		deleted, err := storage.DeleteExpiredCache()
		if err != nil {
			t.Fatalf("failed to delete expired cache: %v", err)
		}
		if deleted == 0 {
			t.Error("expected at least one cache entry to be deleted")
		}
	})
}

func TestConcurrentAccess(t *testing.T) {
	storage, cleanup := setupTestDB(t)
	defer cleanup()

	t.Run("ConcurrentUserCreation", func(t *testing.T) {
		done := make(chan bool)
		for i := 0; i < 10; i++ {
			go func(idx int) {
				user := &User{
					ENumber: "e" + string(rune('0'+idx)),
				}
				storage.CreateUser(user)
				done <- true
			}(i)
		}

		for i := 0; i < 10; i++ {
			<-done
		}

		users, err := storage.ListUsers()
		if err != nil {
			t.Fatalf("failed to list users: %v", err)
		}
		if len(users) < 1 {
			t.Error("expected at least 1 user after concurrent creation")
		}
	})
}
