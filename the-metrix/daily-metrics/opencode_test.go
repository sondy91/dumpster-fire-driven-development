package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCalculateTimeSaved(t *testing.T) {
	tests := []struct {
		name       string
		totalChars int
		wpm        int
		expected   float64
	}{
		{
			name:       "1000 chars at 50 WPM",
			totalChars: 1000,
			wpm:        50,
			expected:   4.0,
		},
		{
			name:       "500 chars at 40 WPM",
			totalChars: 500,
			wpm:        40,
			expected:   2.5,
		},
		{
			name:       "0 chars",
			totalChars: 0,
			wpm:        50,
			expected:   0.0,
		},
		{
			name:       "invalid WPM defaults to 50",
			totalChars: 1000,
			wpm:        0,
			expected:   4.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateTimeSaved(tt.totalChars, tt.wpm)
			if result != tt.expected {
				t.Errorf("calculateTimeSaved(%d, %d) = %.2f; want %.2f",
					tt.totalChars, tt.wpm, result, tt.expected)
			}
		})
	}
}

func TestFormatNumber(t *testing.T) {
	tests := []struct {
		input    int
		expected string
	}{
		{0, "0"},
		{999, "999"},
		{1000, "1,000"},
		{1234567, "1,234,567"},
		{100, "100"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := formatNumber(tt.input)
			if result != tt.expected {
				t.Errorf("formatNumber(%d) = %s; want %s", tt.input, result, tt.expected)
			}
		})
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		minutes  float64
		expected string
	}{
		{0.5, "30 seconds"},
		{1.0, "1.0 minutes"},
		{30.5, "30.5 minutes"},
		{60.0, "1.0 hours"},
		{125.5, "2.1 hours"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := formatDuration(tt.minutes)
			if result != tt.expected {
				t.Errorf("formatDuration(%.1f) = %s; want %s", tt.minutes, result, tt.expected)
			}
		})
	}
}

func TestExportJSON(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "stats.json")

	stats := &OpenCodeStats{
		TotalCommands:    100,
		TotalChars:       5000,
		TimeSavedMins:    20.0,
		SessionCount:     5,
		StartDate:        "Apr 10",
		EndDate:          "Apr 17, 2026",
		TotalEdits:       50,
		TotalCharsEdited: 2000,
		EditTimeSavedMin: 8.0,
		ProjectBreakdown: []ProjectStats{
			{ProjectPath: "/home/user/project1", Commands: 50, Chars: 2500, TimeSavedMins: 10.0},
			{ProjectPath: "/home/user/project2", Commands: 50, Chars: 2500, TimeSavedMins: 10.0},
		},
	}

	err := exportJSON(stats, outputPath)
	if err != nil {
		t.Fatalf("exportJSON failed: %v", err)
	}

	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read exported file: %v", err)
	}

	var result OpenCodeStats
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if result.TotalCommands != stats.TotalCommands {
		t.Errorf("TotalCommands = %d; want %d", result.TotalCommands, stats.TotalCommands)
	}
	if result.TotalEdits != stats.TotalEdits {
		t.Errorf("TotalEdits = %d; want %d", result.TotalEdits, stats.TotalEdits)
	}
	if len(result.ProjectBreakdown) != len(stats.ProjectBreakdown) {
		t.Errorf("ProjectBreakdown length = %d; want %d", len(result.ProjectBreakdown), len(stats.ProjectBreakdown))
	}
}

func TestExportMarkdown(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "report.md")

	stats := &OpenCodeStats{
		TotalCommands:    100,
		TotalChars:       5000,
		TimeSavedMins:    20.0,
		SessionCount:     5,
		StartDate:        "Apr 10",
		EndDate:          "Apr 17, 2026",
		TotalEdits:       50,
		TotalCharsEdited: 2000,
		EditTimeSavedMin: 8.0,
		ProjectBreakdown: []ProjectStats{
			{ProjectPath: "/home/user/project1", Commands: 50, Chars: 2500, TimeSavedMins: 10.0},
		},
	}

	err := exportMarkdown(stats, 50, outputPath)
	if err != nil {
		t.Fatalf("exportMarkdown failed: %v", err)
	}

	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read exported file: %v", err)
	}

	content := string(data)

	expectedSubstrings := []string{
		"# OpenCode AI Time Savings",
		"**Date Range:** Apr 10 - Apr 17, 2026",
		"## Summary",
		"**Sessions analyzed:** 5",
		"**Bash commands:** 100",
		"**File edits:** 50",
		"## Time Savings",
		"**Typing speed:** 50 WPM",
		"## Project Breakdown",
		"| Project | Time Saved | Commands |",
	}

	for _, substring := range expectedSubstrings {
		if !strings.Contains(content, substring) {
			t.Errorf("Markdown output missing substring: %q", substring)
		}
	}
}
