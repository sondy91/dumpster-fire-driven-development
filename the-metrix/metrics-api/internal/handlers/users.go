package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/themetrix/metrics-api/internal/storage"
)

type UserResponse struct {
	ENumber        string  `json:"e_number"`
	JiraAccountID  *string `json:"jira_account_id,omitempty"`
	GitHubUsername *string `json:"github_username,omitempty"`
	Name           *string `json:"name,omitempty"`
	Email          *string `json:"email,omitempty"`
	CreatedAt      string  `json:"created_at"`
	UpdatedAt      string  `json:"updated_at"`
}

type CreateUserRequest struct {
	ENumber        string  `json:"e_number"`
	JiraAccountID  *string `json:"jira_account_id,omitempty"`
	GitHubUsername *string `json:"github_username,omitempty"`
	Name           *string `json:"name,omitempty"`
	Email          *string `json:"email,omitempty"`
}

func (s *Server) ListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := s.Storage.ListUsers()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "database_error", err.Error())
		return
	}

	response := make([]UserResponse, 0, len(users))
	for _, user := range users {
		response = append(response, UserResponse{
			ENumber:        user.ENumber,
			JiraAccountID:  user.JiraAccountID,
			GitHubUsername: user.GitHubUsername,
			Name:           user.Name,
			Email:          user.Email,
			CreatedAt:      user.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:      user.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	respondJSON(w, http.StatusOK, response)
}

func (s *Server) GetUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	eNumber := vars["e_number"]

	user, err := s.Storage.GetUser(eNumber)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "database_error", err.Error())
		return
	}

	if user == nil {
		respondError(w, http.StatusNotFound, "not_found", "user not found")
		return
	}

	response := UserResponse{
		ENumber:        user.ENumber,
		JiraAccountID:  user.JiraAccountID,
		GitHubUsername: user.GitHubUsername,
		Name:           user.Name,
		Email:          user.Email,
		CreatedAt:      user.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:      user.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	respondJSON(w, http.StatusOK, response)
}

func (s *Server) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid_request", "invalid JSON body")
		return
	}

	if req.ENumber == "" {
		respondError(w, http.StatusBadRequest, "validation_error", "e_number is required")
		return
	}

	user := &storage.User{
		ENumber:        req.ENumber,
		JiraAccountID:  req.JiraAccountID,
		GitHubUsername: req.GitHubUsername,
		Name:           req.Name,
		Email:          req.Email,
	}

	if err := s.Storage.CreateUser(user); err != nil {
		respondError(w, http.StatusInternalServerError, "database_error", err.Error())
		return
	}

	created, err := s.Storage.GetUser(req.ENumber)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "database_error", err.Error())
		return
	}

	response := UserResponse{
		ENumber:        created.ENumber,
		JiraAccountID:  created.JiraAccountID,
		GitHubUsername: created.GitHubUsername,
		Name:           created.Name,
		Email:          created.Email,
		CreatedAt:      created.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:      created.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	respondJSON(w, http.StatusCreated, response)
}

func (s *Server) UpdateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	eNumber := vars["e_number"]

	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid_request", "invalid JSON body")
		return
	}

	existing, err := s.Storage.GetUser(eNumber)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "database_error", err.Error())
		return
	}

	if existing == nil {
		respondError(w, http.StatusNotFound, "not_found", "user not found")
		return
	}

	user := &storage.User{
		ENumber:        eNumber,
		JiraAccountID:  req.JiraAccountID,
		GitHubUsername: req.GitHubUsername,
		Name:           req.Name,
		Email:          req.Email,
	}

	if err := s.Storage.UpdateUser(user); err != nil {
		respondError(w, http.StatusInternalServerError, "database_error", err.Error())
		return
	}

	updated, err := s.Storage.GetUser(eNumber)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "database_error", err.Error())
		return
	}

	response := UserResponse{
		ENumber:        updated.ENumber,
		JiraAccountID:  updated.JiraAccountID,
		GitHubUsername: updated.GitHubUsername,
		Name:           updated.Name,
		Email:          updated.Email,
		CreatedAt:      updated.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:      updated.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	respondJSON(w, http.StatusOK, response)
}

func (s *Server) DeleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	eNumber := vars["e_number"]

	existing, err := s.Storage.GetUser(eNumber)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "database_error", err.Error())
		return
	}

	if existing == nil {
		respondError(w, http.StatusNotFound, "not_found", "user not found")
		return
	}

	if err := s.Storage.DeleteUser(eNumber); err != nil {
		respondError(w, http.StatusInternalServerError, "database_error", err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
