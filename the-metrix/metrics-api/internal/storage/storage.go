package storage

import (
	"database/sql"
	"embed"
	"fmt"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed migrations/*.sql
var migrations embed.FS

type Storage struct {
	db *sql.DB
}

type User struct {
	ID             int
	ENumber        string
	JiraAccountID  *string
	GitHubUsername *string
	Name           *string
	Email          *string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type Repo struct {
	ID        int
	Name      string
	Path      string
	RepoType  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ConfigEntry struct {
	Key         string
	Value       string
	Description *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type MetricCache struct {
	ID         int
	CacheKey   string
	MetricType string
	ENumber    *string
	StartDate  *time.Time
	EndDate    *time.Time
	Data       string
	CreatedAt  time.Time
	ExpiresAt  time.Time
}

type SurveyResponse struct {
	ID                  string
	SubmittedAt         string
	Name                string
	Sprint              string
	Flow                *int
	AISatisfaction      *int
	AISpeed             *int
	AICodePct           *int
	DevEx               *int
	Blockers            string
	Improvement         *string
	AITool              *string
	SpacePerformance    *int
	SpaceCollaboration  *int
	SpaceActivity       *int
	SpaceEfficiency     *int
	CreatedAt           time.Time
}

type OpenCodeSession struct {
	ID                 string
	SavedAt            string
	Developer          string
	Project            string
	Sprint             string
	Bash               int
	Reads              int
	Edits              int
	Searches           int
	TimeSavedBash      float64
	TimeSavedReads     float64
	TimeSavedEdits     float64
	TimeSavedSearches  float64
	TimeSavedTotal     float64
	SessionDuration    int
	RawText            *string
	CreatedAt          time.Time
}

type OpenCodeActivity struct {
	ID                  string
	Date                string
	Developer           string
	ProjectPath         string
	ProjectName         *string
	BashCommands        int
	FileReads           int
	FileEdits           int
	FileWrites          int
	Searches            int
	TimeSavedBash       float64
	TimeSavedReads      float64
	TimeSavedEdits      float64
	TimeSavedWrites     float64
	TimeSavedSearches   float64
	TimeSavedTotal      float64
	AvgComplexityScore  float64
	SessionCount        int
	CharsTyped          int
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

type Feedback struct {
	ID        string
	Timestamp string
	Page      string
	User      *string
	UserAgent string
	Kind      string
	Type      *string
	Text      *string
	Component *string
	Sentiment *string
	CreatedAt time.Time
}

func New(dbPath string) (*Storage, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	storage := &Storage{db: db}

	if err := storage.runMigrations(); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return storage, nil
}

func (s *Storage) Close() error {
	return s.db.Close()
}

func (s *Storage) runMigrations() error {
	migrationFiles := []string{
		"migrations/001_initial_schema.sql",
		"migrations/002_survey_opencode_tables.sql",
		"migrations/007_create_feedback_table.sql",
		"migrations/008_opencode_activity_table.sql",
	}

	for _, file := range migrationFiles {
		schema, err := migrations.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", file, err)
		}

		if _, err := s.db.Exec(string(schema)); err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", file, err)
		}
	}

	return nil
}

// User operations
func (s *Storage) CreateUser(user *User) error {
	query := `
		INSERT INTO users (e_number, jira_account_id, github_username, name, email)
		VALUES (?, ?, ?, ?, ?)
	`
	_, err := s.db.Exec(query, user.ENumber, user.JiraAccountID, user.GitHubUsername, user.Name, user.Email)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

func (s *Storage) GetUser(eNumber string) (*User, error) {
	query := `
		SELECT id, e_number, jira_account_id, github_username, name, email, created_at, updated_at
		FROM users
		WHERE e_number = ?
	`
	var user User
	err := s.db.QueryRow(query, eNumber).Scan(
		&user.ID,
		&user.ENumber,
		&user.JiraAccountID,
		&user.GitHubUsername,
		&user.Name,
		&user.Email,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &user, nil
}

func (s *Storage) UpdateUser(user *User) error {
	query := `
		UPDATE users
		SET jira_account_id = ?, github_username = ?, name = ?, email = ?
		WHERE e_number = ?
	`
	result, err := s.db.Exec(query, user.JiraAccountID, user.GitHubUsername, user.Name, user.Email, user.ENumber)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("user not found: %s", user.ENumber)
	}

	return nil
}

func (s *Storage) DeleteUser(eNumber string) error {
	query := `DELETE FROM users WHERE e_number = ?`
	result, err := s.db.Exec(query, eNumber)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("user not found: %s", eNumber)
	}

	return nil
}

func (s *Storage) ListUsers() ([]*User, error) {
	query := `
		SELECT id, e_number, jira_account_id, github_username, name, email, created_at, updated_at
		FROM users
		ORDER BY e_number
	`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		var user User
		err := rows.Scan(
			&user.ID,
			&user.ENumber,
			&user.JiraAccountID,
			&user.GitHubUsername,
			&user.Name,
			&user.Email,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate users: %w", err)
	}

	return users, nil
}

// Repo operations
func (s *Storage) CreateRepo(repo *Repo) error {
	query := `
		INSERT INTO repos (name, path, repo_type)
		VALUES (?, ?, ?)
	`
	result, err := s.db.Exec(query, repo.Name, repo.Path, repo.RepoType)
	if err != nil {
		return fmt.Errorf("failed to create repo: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	repo.ID = int(id)
	return nil
}

func (s *Storage) GetRepo(path string) (*Repo, error) {
	query := `
		SELECT id, name, path, repo_type, created_at, updated_at
		FROM repos
		WHERE path = ?
	`
	var repo Repo
	err := s.db.QueryRow(query, path).Scan(
		&repo.ID,
		&repo.Name,
		&repo.Path,
		&repo.RepoType,
		&repo.CreatedAt,
		&repo.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get repo: %w", err)
	}
	return &repo, nil
}

func (s *Storage) ListRepos() ([]*Repo, error) {
	query := `
		SELECT id, name, path, repo_type, created_at, updated_at
		FROM repos
		ORDER BY name
	`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list repos: %w", err)
	}
	defer rows.Close()

	var repos []*Repo
	for rows.Next() {
		var repo Repo
		err := rows.Scan(
			&repo.ID,
			&repo.Name,
			&repo.Path,
			&repo.RepoType,
			&repo.CreatedAt,
			&repo.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan repo: %w", err)
		}
		repos = append(repos, &repo)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate repos: %w", err)
	}

	return repos, nil
}

// Config operations
func (s *Storage) SetConfig(key, value string) error {
	query := `
		INSERT INTO config (key, value) VALUES (?, ?)
		ON CONFLICT(key) DO UPDATE SET value = ?, updated_at = CURRENT_TIMESTAMP
	`
	_, err := s.db.Exec(query, key, value, value)
	if err != nil {
		return fmt.Errorf("failed to set config: %w", err)
	}
	return nil
}

func (s *Storage) GetConfig(key string) (string, error) {
	query := `SELECT value FROM config WHERE key = ?`
	var value string
	err := s.db.QueryRow(query, key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("failed to get config: %w", err)
	}
	return value, nil
}

// Cache operations
func (s *Storage) SetCache(metricType, cacheKey, data string, eNumber *string, startDate, endDate *time.Time, ttl time.Duration) error {
	expiresAt := time.Now().Add(ttl)
	query := `
		INSERT INTO metric_cache (cache_key, metric_type, e_number, start_date, end_date, data, expires_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(cache_key) DO UPDATE SET data = ?, expires_at = ?, created_at = CURRENT_TIMESTAMP
	`
	_, err := s.db.Exec(query, cacheKey, metricType, eNumber, startDate, endDate, data, expiresAt, data, expiresAt)
	if err != nil {
		return fmt.Errorf("failed to set cache: %w", err)
	}
	return nil
}

func (s *Storage) GetCache(cacheKey string) (string, bool, error) {
	query := `SELECT data, expires_at FROM metric_cache WHERE cache_key = ?`
	var data string
	var expiresAt time.Time
	err := s.db.QueryRow(query, cacheKey).Scan(&data, &expiresAt)
	if err == sql.ErrNoRows {
		return "", false, nil
	}
	if err != nil {
		return "", false, fmt.Errorf("failed to get cache: %w", err)
	}

	if time.Now().After(expiresAt) {
		return "", false, nil
	}

	return data, true, nil
}

func (s *Storage) DeleteExpiredCache() (int64, error) {
	query := `DELETE FROM metric_cache WHERE expires_at < CURRENT_TIMESTAMP`
	result, err := s.db.Exec(query)
	if err != nil {
		return 0, fmt.Errorf("failed to delete expired cache: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rows, nil
}

// Survey operations
func (s *Storage) CreateSurveyResponse(resp *SurveyResponse) error {
	query := `
		INSERT INTO survey_responses (
			id, submitted_at, name, sprint, flow, ai_satisfaction, ai_speed,
			ai_code_pct, devex, blockers, improvement, ai_tool,
			space_performance, space_collaboration, space_activity, space_efficiency
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := s.db.Exec(query,
		resp.ID, resp.SubmittedAt, resp.Name, resp.Sprint,
		resp.Flow, resp.AISatisfaction, resp.AISpeed, resp.AICodePct,
		resp.DevEx, resp.Blockers, resp.Improvement, resp.AITool,
		resp.SpacePerformance, resp.SpaceCollaboration, resp.SpaceActivity, resp.SpaceEfficiency,
	)
	if err != nil {
		return fmt.Errorf("failed to create survey response: %w", err)
	}
	return nil
}

func (s *Storage) ListSurveyResponses(sprint *string, limit int) ([]*SurveyResponse, error) {
	query := `
		SELECT id, submitted_at, name, sprint, flow, ai_satisfaction, ai_speed,
			ai_code_pct, devex, blockers, improvement, ai_tool,
			space_performance, space_collaboration, space_activity, space_efficiency, created_at
		FROM survey_responses
	`
	args := []interface{}{}
	if sprint != nil {
		query += " WHERE sprint = ?"
		args = append(args, *sprint)
	}
	query += " ORDER BY submitted_at DESC"
	if limit > 0 {
		query += " LIMIT ?"
		args = append(args, limit)
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list survey responses: %w", err)
	}
	defer rows.Close()

	var responses []*SurveyResponse
	for rows.Next() {
		var resp SurveyResponse
		err := rows.Scan(
			&resp.ID, &resp.SubmittedAt, &resp.Name, &resp.Sprint,
			&resp.Flow, &resp.AISatisfaction, &resp.AISpeed, &resp.AICodePct,
			&resp.DevEx, &resp.Blockers, &resp.Improvement, &resp.AITool,
			&resp.SpacePerformance, &resp.SpaceCollaboration, &resp.SpaceActivity, &resp.SpaceEfficiency,
			&resp.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan survey response: %w", err)
		}
		responses = append(responses, &resp)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate survey responses: %w", err)
	}

	return responses, nil
}

func (s *Storage) DeleteAllSurveyResponses() error {
	query := `DELETE FROM survey_responses`
	result, err := s.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to delete survey responses: %w", err)
	}
	
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	log.Printf("Deleted %d survey responses", rows)
	return nil
}

// OpenCode operations
func (s *Storage) CreateOpenCodeSession(sess *OpenCodeSession) error {
	query := `
		INSERT INTO opencode_sessions (
			id, saved_at, developer, project, sprint,
			bash, reads, edits, searches,
			time_saved_bash, time_saved_reads, time_saved_edits, time_saved_searches, time_saved_total,
			session_duration, raw_text
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := s.db.Exec(query,
		sess.ID, sess.SavedAt, sess.Developer, sess.Project, sess.Sprint,
		sess.Bash, sess.Reads, sess.Edits, sess.Searches,
		sess.TimeSavedBash, sess.TimeSavedReads, sess.TimeSavedEdits, sess.TimeSavedSearches, sess.TimeSavedTotal,
		sess.SessionDuration, sess.RawText,
	)
	if err != nil {
		return fmt.Errorf("failed to create opencode session: %w", err)
	}
	return nil
}

func (s *Storage) ListOpenCodeSessions(developer *string, project *string, sprint *string, limit int) ([]*OpenCodeSession, error) {
	query := `
		SELECT id, saved_at, developer, project, sprint,
			bash, reads, edits, searches,
			time_saved_bash, time_saved_reads, time_saved_edits, time_saved_searches, time_saved_total,
			session_duration, raw_text, created_at
		FROM opencode_sessions
		WHERE 1=1
	`
	args := []interface{}{}
	if developer != nil {
		query += " AND developer = ?"
		args = append(args, *developer)
	}
	if project != nil {
		query += " AND project = ?"
		args = append(args, *project)
	}
	if sprint != nil {
		query += " AND sprint = ?"
		args = append(args, *sprint)
	}
	query += " ORDER BY saved_at DESC"
	if limit > 0 {
		query += " LIMIT ?"
		args = append(args, limit)
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list opencode sessions: %w", err)
	}
	defer rows.Close()

	var sessions []*OpenCodeSession
	for rows.Next() {
		var sess OpenCodeSession
		err := rows.Scan(
			&sess.ID, &sess.SavedAt, &sess.Developer, &sess.Project, &sess.Sprint,
			&sess.Bash, &sess.Reads, &sess.Edits, &sess.Searches,
			&sess.TimeSavedBash, &sess.TimeSavedReads, &sess.TimeSavedEdits, &sess.TimeSavedSearches, &sess.TimeSavedTotal,
			&sess.SessionDuration, &sess.RawText, &sess.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan opencode session: %w", err)
		}
		sessions = append(sessions, &sess)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate opencode sessions: %w", err)
	}

	return sessions, nil
}

func (s *Storage) DeleteOpenCodeSession(id string) error {
	query := `DELETE FROM opencode_sessions WHERE id = ?`
	result, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("session not found: %s", id)
	}

	return nil
}

// OpenCodeActivity operations
func (s *Storage) CreateOrUpdateOpenCodeActivity(activity *OpenCodeActivity) error {
	query := `
		INSERT INTO opencode_activity (
			id, date, developer, project_path, project_name,
			bash_commands, file_reads, file_edits, file_writes, searches,
			time_saved_bash, time_saved_reads, time_saved_edits, time_saved_writes, time_saved_searches, time_saved_total,
			avg_complexity_score, session_count, chars_typed
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(date, developer, project_path) DO UPDATE SET
			project_name = excluded.project_name,
			bash_commands = excluded.bash_commands,
			file_reads = excluded.file_reads,
			file_edits = excluded.file_edits,
			file_writes = excluded.file_writes,
			searches = excluded.searches,
			time_saved_bash = excluded.time_saved_bash,
			time_saved_reads = excluded.time_saved_reads,
			time_saved_edits = excluded.time_saved_edits,
			time_saved_writes = excluded.time_saved_writes,
			time_saved_searches = excluded.time_saved_searches,
			time_saved_total = excluded.time_saved_total,
			avg_complexity_score = excluded.avg_complexity_score,
			session_count = excluded.session_count,
			chars_typed = excluded.chars_typed,
			updated_at = CURRENT_TIMESTAMP
	`
	_, err := s.db.Exec(query,
		activity.ID, activity.Date, activity.Developer, activity.ProjectPath, activity.ProjectName,
		activity.BashCommands, activity.FileReads, activity.FileEdits, activity.FileWrites, activity.Searches,
		activity.TimeSavedBash, activity.TimeSavedReads, activity.TimeSavedEdits, activity.TimeSavedWrites, activity.TimeSavedSearches, activity.TimeSavedTotal,
		activity.AvgComplexityScore, activity.SessionCount, activity.CharsTyped,
	)
	if err != nil {
		return fmt.Errorf("failed to create or update opencode activity: %w", err)
	}
	return nil
}

func (s *Storage) ListOpenCodeActivity(startDate, endDate *string, developer, projectName *string, limit int) ([]*OpenCodeActivity, error) {
	query := `
		SELECT id, date, developer, project_path, project_name,
			bash_commands, file_reads, file_edits, file_writes, searches,
			time_saved_bash, time_saved_reads, time_saved_edits, time_saved_writes, time_saved_searches, time_saved_total,
			avg_complexity_score, session_count, chars_typed, created_at, updated_at
		FROM opencode_activity
		WHERE 1=1
	`
	args := []interface{}{}

	if startDate != nil {
		query += " AND date >= ?"
		args = append(args, *startDate)
	}
	if endDate != nil {
		query += " AND date <= ?"
		args = append(args, *endDate)
	}
	if developer != nil {
		query += " AND developer = ?"
		args = append(args, *developer)
	}
	if projectName != nil {
		query += " AND project_name = ?"
		args = append(args, *projectName)
	}

	query += " ORDER BY date DESC, time_saved_total DESC LIMIT ?"
	args = append(args, limit)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list opencode activity: %w", err)
	}
	defer rows.Close()

	var activities []*OpenCodeActivity
	for rows.Next() {
		var activity OpenCodeActivity
		err := rows.Scan(
			&activity.ID, &activity.Date, &activity.Developer, &activity.ProjectPath, &activity.ProjectName,
			&activity.BashCommands, &activity.FileReads, &activity.FileEdits, &activity.FileWrites, &activity.Searches,
			&activity.TimeSavedBash, &activity.TimeSavedReads, &activity.TimeSavedEdits, &activity.TimeSavedWrites, &activity.TimeSavedSearches, &activity.TimeSavedTotal,
			&activity.AvgComplexityScore, &activity.SessionCount, &activity.CharsTyped, &activity.CreatedAt, &activity.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan opencode activity: %w", err)
		}
		activities = append(activities, &activity)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate opencode activity: %w", err)
	}

	return activities, nil
}

func (s *Storage) CreateFeedback(feedback *Feedback) error {
	query := `
		INSERT INTO feedback (
			id, timestamp, page, user, user_agent, kind, type, text, component, sentiment, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	`
	_, err := s.db.Exec(
		query,
		feedback.ID,
		feedback.Timestamp,
		feedback.Page,
		feedback.User,
		feedback.UserAgent,
		feedback.Kind,
		feedback.Type,
		feedback.Text,
		feedback.Component,
		feedback.Sentiment,
	)
	return err
}

func (s *Storage) ListFeedback(page string, kind string, limit int) ([]*Feedback, error) {
	query := `SELECT id, timestamp, page, user, user_agent, kind, type, text, component, sentiment, created_at
			  FROM feedback WHERE 1=1`
	args := []interface{}{}

	if page != "" {
		query += " AND page = ?"
		args = append(args, page)
	}
	if kind != "" {
		query += " AND kind = ?"
		args = append(args, kind)
	}

	query += " ORDER BY created_at DESC LIMIT ?"
	args = append(args, limit)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query feedback: %w", err)
	}
	defer rows.Close()

	var feedback []*Feedback
	for rows.Next() {
		var f Feedback
		err := rows.Scan(
			&f.ID, &f.Timestamp, &f.Page, &f.User, &f.UserAgent,
			&f.Kind, &f.Type, &f.Text, &f.Component, &f.Sentiment, &f.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan feedback: %w", err)
		}
		feedback = append(feedback, &f)
	}

	return feedback, nil
}
