package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/themetrix/metrics-api/internal/storage"
)

type OpenCodeHandler struct {
	storage *storage.Storage
}

func NewOpenCodeHandler(s *storage.Storage) *OpenCodeHandler {
	return &OpenCodeHandler{storage: s}
}

type OpenCodeSubmitRequest struct {
	ID                 string   `json:"id"`
	SavedAt            string   `json:"saved_at"`
	Developer          string   `json:"developer"`
	Project            string   `json:"project"`
	Sprint             string   `json:"sprint"`
	Bash               int      `json:"bash"`
	Reads              int      `json:"reads"`
	Edits              int      `json:"edits"`
	Searches           int      `json:"searches"`
	TimeSavedBash      float64  `json:"time_saved_bash"`
	TimeSavedReads     float64  `json:"time_saved_reads"`
	TimeSavedEdits     float64  `json:"time_saved_edits"`
	TimeSavedSearches  float64  `json:"time_saved_searches"`
	TimeSavedTotal     float64  `json:"time_saved_total"`
	SessionDuration    int      `json:"session_duration"`
	RawText            *string  `json:"raw_text"`
}

func (h *OpenCodeHandler) SubmitSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req OpenCodeSubmitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	// Generate ID if not provided
	if req.ID == "" {
		req.ID = fmt.Sprintf("%d", time.Now().UnixNano())
	}

	// Validate required fields
	if req.Developer == "" || req.Project == "" || req.Sprint == "" {
		http.Error(w, "developer, project, and sprint are required", http.StatusBadRequest)
		return
	}

	sess := &storage.OpenCodeSession{
		ID:                 req.ID,
		SavedAt:            req.SavedAt,
		Developer:          req.Developer,
		Project:            req.Project,
		Sprint:             req.Sprint,
		Bash:               req.Bash,
		Reads:              req.Reads,
		Edits:              req.Edits,
		Searches:           req.Searches,
		TimeSavedBash:      req.TimeSavedBash,
		TimeSavedReads:     req.TimeSavedReads,
		TimeSavedEdits:     req.TimeSavedEdits,
		TimeSavedSearches:  req.TimeSavedSearches,
		TimeSavedTotal:     req.TimeSavedTotal,
		SessionDuration:    req.SessionDuration,
		RawText:            req.RawText,
	}

	if err := h.storage.CreateOpenCodeSession(sess); err != nil {
		http.Error(w, fmt.Sprintf("Failed to save session: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"id":      sess.ID,
	})
}

func (h *OpenCodeHandler) GetSessions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	query := r.URL.Query()
	var developer, project, sprint *string
	if d := query.Get("developer"); d != "" {
		developer = &d
	}
	if p := query.Get("project"); p != "" {
		project = &p
	}
	if s := query.Get("sprint"); s != "" {
		sprint = &s
	}

	limit := 100
	if l := query.Get("limit"); l != "" {
		if parsedLimit, err := strconv.Atoi(l); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	sessions, err := h.storage.ListOpenCodeSessions(developer, project, sprint, limit)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to fetch sessions: %v", err), http.StatusInternalServerError)
		return
	}

	// Convert to API format
	apiSessions := make([]map[string]interface{}, 0, len(sessions))
	for _, sess := range sessions {
		apiSess := map[string]interface{}{
			"id":                   sess.ID,
			"saved_at":             sess.SavedAt,
			"developer":            sess.Developer,
			"project":              sess.Project,
			"sprint":               sess.Sprint,
			"bash":                 sess.Bash,
			"reads":                sess.Reads,
			"edits":                sess.Edits,
			"searches":             sess.Searches,
			"time_saved_bash":      sess.TimeSavedBash,
			"time_saved_reads":     sess.TimeSavedReads,
			"time_saved_edits":     sess.TimeSavedEdits,
			"time_saved_searches":  sess.TimeSavedSearches,
			"time_saved_total":     sess.TimeSavedTotal,
			"session_duration":     sess.SessionDuration,
		}
		apiSessions = append(apiSessions, apiSess)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"count":    len(apiSessions),
		"sessions": apiSessions,
	})
}

func (h *OpenCodeHandler) DeleteSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sessionID := r.URL.Query().Get("id")
	if sessionID == "" {
		http.Error(w, "Session ID is required", http.StatusBadRequest)
		return
	}

	if err := h.storage.DeleteOpenCodeSession(sessionID); err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete session: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Session deleted successfully",
		"id":      sessionID,
	})
}

// OpenCodeActivity handlers

type OpenCodeActivitySubmitRequest struct {
	ID                 string   `json:"id"`
	Date               string   `json:"date"`
	Developer          string   `json:"developer"`
	ProjectPath        string   `json:"project_path"`
	ProjectName        *string  `json:"project_name"`
	BashCommands       int      `json:"bash_commands"`
	FileReads          int      `json:"file_reads"`
	FileEdits          int      `json:"file_edits"`
	FileWrites         int      `json:"file_writes"`
	Searches           int      `json:"searches"`
	TimeSavedBash      float64  `json:"time_saved_bash"`
	TimeSavedReads     float64  `json:"time_saved_reads"`
	TimeSavedEdits     float64  `json:"time_saved_edits"`
	TimeSavedWrites    float64  `json:"time_saved_writes"`
	TimeSavedSearches  float64  `json:"time_saved_searches"`
	TimeSavedTotal     float64  `json:"time_saved_total"`
	AvgComplexityScore float64  `json:"avg_complexity_score"`
	SessionCount       int      `json:"session_count"`
	CharsTyped         int      `json:"chars_typed"`
}

func (h *OpenCodeHandler) SubmitActivity(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req OpenCodeActivitySubmitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Date == "" || req.Developer == "" || req.ProjectPath == "" {
		http.Error(w, "date, developer, and project_path are required", http.StatusBadRequest)
		return
	}

	// Generate ID if not provided
	if req.ID == "" {
		req.ID = fmt.Sprintf("%s-%s-%s", req.Date, req.Developer, req.ProjectPath)
	}

	activity := &storage.OpenCodeActivity{
		ID:                  req.ID,
		Date:                req.Date,
		Developer:           req.Developer,
		ProjectPath:         req.ProjectPath,
		ProjectName:         req.ProjectName,
		BashCommands:        req.BashCommands,
		FileReads:           req.FileReads,
		FileEdits:           req.FileEdits,
		FileWrites:          req.FileWrites,
		Searches:            req.Searches,
		TimeSavedBash:       req.TimeSavedBash,
		TimeSavedReads:      req.TimeSavedReads,
		TimeSavedEdits:      req.TimeSavedEdits,
		TimeSavedWrites:     req.TimeSavedWrites,
		TimeSavedSearches:   req.TimeSavedSearches,
		TimeSavedTotal:      req.TimeSavedTotal,
		AvgComplexityScore:  req.AvgComplexityScore,
		SessionCount:        req.SessionCount,
		CharsTyped:          req.CharsTyped,
	}

	if err := h.storage.CreateOrUpdateOpenCodeActivity(activity); err != nil {
		http.Error(w, fmt.Sprintf("Failed to save activity: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"id":      activity.ID,
	})
}

func (h *OpenCodeHandler) GetActivity(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	query := r.URL.Query()
	var startDate, endDate, developer, projectName *string
	
	if d := query.Get("start_date"); d != "" {
		startDate = &d
	}
	if d := query.Get("end_date"); d != "" {
		endDate = &d
	}
	if d := query.Get("developer"); d != "" {
		developer = &d
	}
	if p := query.Get("project_name"); p != "" {
		projectName = &p
	}

	limit := 1000
	if l := query.Get("limit"); l != "" {
		if parsedLimit, err := strconv.Atoi(l); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	activities, err := h.storage.ListOpenCodeActivity(startDate, endDate, developer, projectName, limit)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to fetch activities: %v", err), http.StatusInternalServerError)
		return
	}

	// Convert to API format
	apiActivities := make([]map[string]interface{}, 0, len(activities))
	for _, act := range activities {
		apiAct := map[string]interface{}{
			"id":                    act.ID,
			"date":                  act.Date,
			"developer":             act.Developer,
			"project_path":          act.ProjectPath,
			"project_name":          act.ProjectName,
			"bash_commands":         act.BashCommands,
			"file_reads":            act.FileReads,
			"file_edits":            act.FileEdits,
			"file_writes":           act.FileWrites,
			"searches":              act.Searches,
			"time_saved_bash":       act.TimeSavedBash,
			"time_saved_reads":      act.TimeSavedReads,
			"time_saved_edits":      act.TimeSavedEdits,
			"time_saved_writes":     act.TimeSavedWrites,
			"time_saved_searches":   act.TimeSavedSearches,
			"time_saved_total":      act.TimeSavedTotal,
			"avg_complexity_score":  act.AvgComplexityScore,
			"session_count":         act.SessionCount,
			"chars_typed":           act.CharsTyped,
			"created_at":            act.CreatedAt,
			"updated_at":            act.UpdatedAt,
		}
		apiActivities = append(apiActivities, apiAct)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"count":      len(apiActivities),
		"activities": apiActivities,
	})
}
