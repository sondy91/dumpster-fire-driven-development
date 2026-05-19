package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/themetrix/metrics-api/internal/storage"
)

type FeedbackHandler struct {
	storage *storage.Storage
}

func NewFeedbackHandler(s *storage.Storage) *FeedbackHandler {
	return &FeedbackHandler{storage: s}
}

type FeedbackSubmitRequest struct {
	ID        string                 `json:"id"`
	Timestamp string                 `json:"timestamp"`
	Page      string                 `json:"page"`
	User      *string                `json:"user"`
	UserAgent string                 `json:"userAgent"`
	Kind      string                 `json:"kind"` // "widget" or "reaction"
	Type      *string                `json:"type"` // For widget: "feature", "bug", "idea", "broken"
	Text      *string                `json:"text"`
	Component *string                `json:"component"` // For reaction
	Sentiment *string                `json:"sentiment"` // For reaction: "up" or "down"
}

func (h *FeedbackHandler) SubmitFeedback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req FeedbackSubmitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.ID == "" || req.Page == "" || req.Kind == "" {
		http.Error(w, "id, page, and kind are required", http.StatusBadRequest)
		return
	}

	feedback := &storage.Feedback{
		ID:        req.ID,
		Timestamp: req.Timestamp,
		Page:      req.Page,
		User:      req.User,
		UserAgent: req.UserAgent,
		Kind:      req.Kind,
		Type:      req.Type,
		Text:      req.Text,
		Component: req.Component,
		Sentiment: req.Sentiment,
	}

	if err := h.storage.CreateFeedback(feedback); err != nil {
		http.Error(w, fmt.Sprintf("Failed to save feedback: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "success",
		"id":     req.ID,
	})
}

func (h *FeedbackHandler) GetFeedback(w http.ResponseWriter, r *http.Request) {
	page := r.URL.Query().Get("page")
	kind := r.URL.Query().Get("kind")
	limit := 100

	feedback, err := h.storage.ListFeedback(page, kind, limit)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list feedback: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"count":    len(feedback),
		"feedback": feedback,
	})
}
