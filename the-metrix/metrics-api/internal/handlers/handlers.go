package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/themetrix/metrics-api/internal/storage"
	"github.com/themetrix/metrics-api/pkg/config"
)

type Server struct {
	Storage *storage.Storage
	Config  *config.MetricsConfig
}

func NewServer(store *storage.Storage, cfg *config.MetricsConfig) *Server {
	return &Server{
		Storage: store,
		Config:  cfg,
	}
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

type HealthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
	Version string `json:"version"`
}

func (s *Server) Health(w http.ResponseWriter, r *http.Request) {
	response := HealthResponse{
		Status:  "healthy",
		Service: "metrics-api",
		Version: "0.1.0",
	}
	respondJSON(w, http.StatusOK, response)
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, err string, message string) {
	response := ErrorResponse{
		Error:   err,
		Message: message,
	}
	respondJSON(w, status, response)
}
