package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/themetrix/metrics-api/internal/storage"
)

type SurveyHandler struct {
	storage *storage.Storage
}

func NewSurveyHandler(s *storage.Storage) *SurveyHandler {
	return &SurveyHandler{storage: s}
}

type SurveySubmitRequest struct {
	ID                  string   `json:"id"`
	SubmittedAt         string   `json:"submitted_at"`
	Name                string   `json:"name"`
	Sprint              string   `json:"sprint"`
	Flow                *int     `json:"flow"`
	AISatisfaction      *int     `json:"ai_satisfaction"`
	AISpeed             *int     `json:"ai_speed"`
	AICodePct           *int     `json:"ai_code_pct"`
	DevEx               *int     `json:"devex"`
	Blockers            []string `json:"blockers"`
	Improvement         *string  `json:"improvement"`
	AITool              *string  `json:"ai_tool"`
	SpacePerformance    *int     `json:"space_performance"`
	SpaceCollaboration  *int     `json:"space_collaboration"`
	SpaceActivity       *int     `json:"space_activity"`
	SpaceEfficiency     *int     `json:"space_efficiency"`
}

func (h *SurveyHandler) SubmitSurvey(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SurveySubmitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	// Generate ID if not provided
	if req.ID == "" {
		req.ID = fmt.Sprintf("%d", time.Now().UnixNano())
	}

	// Validate required fields
	if req.Sprint == "" {
		http.Error(w, "sprint is required", http.StatusBadRequest)
		return
	}

	// Convert blockers array to JSON string
	blockersJSON, err := json.Marshal(req.Blockers)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode blockers: %v", err), http.StatusInternalServerError)
		return
	}

	resp := &storage.SurveyResponse{
		ID:                  req.ID,
		SubmittedAt:         req.SubmittedAt,
		Name:                req.Name,
		Sprint:              req.Sprint,
		Flow:                req.Flow,
		AISatisfaction:      req.AISatisfaction,
		AISpeed:             req.AISpeed,
		AICodePct:           req.AICodePct,
		DevEx:               req.DevEx,
		Blockers:            string(blockersJSON),
		Improvement:         req.Improvement,
		AITool:              req.AITool,
		SpacePerformance:    req.SpacePerformance,
		SpaceCollaboration:  req.SpaceCollaboration,
		SpaceActivity:       req.SpaceActivity,
		SpaceEfficiency:     req.SpaceEfficiency,
	}

	if err := h.storage.CreateSurveyResponse(resp); err != nil {
		http.Error(w, fmt.Sprintf("Failed to save survey response: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"id":      resp.ID,
	})
}

func (h *SurveyHandler) GetSurveyResponses(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	query := r.URL.Query()
	var sprint *string
	if s := query.Get("sprint"); s != "" {
		sprint = &s
	}

	limit := 100
	if l := query.Get("limit"); l != "" {
		if parsedLimit, err := strconv.Atoi(l); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	responses, err := h.storage.ListSurveyResponses(sprint, limit)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to fetch survey responses: %v", err), http.StatusInternalServerError)
		return
	}

	// Convert responses to API format
	apiResponses := make([]map[string]interface{}, 0, len(responses))
	for _, resp := range responses {
		// Parse blockers JSON
		var blockers []string
		if resp.Blockers != "" {
			json.Unmarshal([]byte(resp.Blockers), &blockers)
		}

		apiResp := map[string]interface{}{
			"id":                    resp.ID,
			"submitted_at":          resp.SubmittedAt,
			"name":                  resp.Name,
			"sprint":                resp.Sprint,
			"flow":                  resp.Flow,
			"ai_satisfaction":       resp.AISatisfaction,
			"ai_speed":              resp.AISpeed,
			"ai_code_pct":           resp.AICodePct,
			"devex":                 resp.DevEx,
			"blockers":              blockers,
			"improvement":           resp.Improvement,
			"ai_tool":               resp.AITool,
			"space_performance":     resp.SpacePerformance,
			"space_collaboration":   resp.SpaceCollaboration,
			"space_activity":        resp.SpaceActivity,
			"space_efficiency":      resp.SpaceEfficiency,
		}
		apiResponses = append(apiResponses, apiResp)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"count":     len(apiResponses),
		"responses": apiResponses,
	})
}

func (h *SurveyHandler) DeleteAllSurveys(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := h.storage.DeleteAllSurveyResponses(); err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete survey responses: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "All survey responses deleted",
	})
}
