package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/mux"
	"github.com/themetrix/metrics-api/internal/storage"
)

type RepoResponse struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Path      string `json:"path"`
	RepoType  string `json:"repo_type"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type CreateRepoRequest struct {
	Name     string `json:"name"`
	Path     string `json:"path"`
	RepoType string `json:"repo_type"`
}

func (s *Server) ListRepos(w http.ResponseWriter, r *http.Request) {
	repos, err := s.Storage.ListRepos()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "database_error", err.Error())
		return
	}

	response := make([]RepoResponse, 0, len(repos))
	for _, repo := range repos {
		response = append(response, RepoResponse{
			ID:        repo.ID,
			Name:      repo.Name,
			Path:      repo.Path,
			RepoType:  repo.RepoType,
			CreatedAt: repo.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt: repo.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	respondJSON(w, http.StatusOK, response)
}

func (s *Server) GetRepo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	path := vars["path"]

	repo, err := s.Storage.GetRepo(path)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "database_error", err.Error())
		return
	}

	if repo == nil {
		respondError(w, http.StatusNotFound, "not_found", "repo not found")
		return
	}

	response := RepoResponse{
		ID:        repo.ID,
		Name:      repo.Name,
		Path:      repo.Path,
		RepoType:  repo.RepoType,
		CreatedAt: repo.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: repo.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	respondJSON(w, http.StatusOK, response)
}

func (s *Server) CreateRepo(w http.ResponseWriter, r *http.Request) {
	var req CreateRepoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid_request", "invalid JSON body")
		return
	}

	if req.Name == "" || req.Path == "" {
		respondError(w, http.StatusBadRequest, "validation_error", "name and path are required")
		return
	}

	if req.RepoType == "" {
		req.RepoType = "work"
	}

	if req.RepoType != "work" && req.RepoType != "private" {
		respondError(w, http.StatusBadRequest, "validation_error", "repo_type must be 'work' or 'private'")
		return
	}

	if _, err := os.Stat(req.Path); os.IsNotExist(err) {
		respondError(w, http.StatusBadRequest, "validation_error", "path does not exist")
		return
	}

	gitPath := filepath.Join(req.Path, ".git")
	if _, err := os.Stat(gitPath); os.IsNotExist(err) {
		respondError(w, http.StatusBadRequest, "validation_error", "path is not a git repository")
		return
	}

	repo := &storage.Repo{
		Name:     req.Name,
		Path:     req.Path,
		RepoType: req.RepoType,
	}

	if err := s.Storage.CreateRepo(repo); err != nil {
		respondError(w, http.StatusInternalServerError, "database_error", err.Error())
		return
	}

	created, err := s.Storage.GetRepo(req.Path)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "database_error", err.Error())
		return
	}

	response := RepoResponse{
		ID:        created.ID,
		Name:      created.Name,
		Path:      created.Path,
		RepoType:  created.RepoType,
		CreatedAt: created.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: created.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	respondJSON(w, http.StatusCreated, response)
}
