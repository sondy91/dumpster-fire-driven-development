package calendar

import (
	"testing"
	"time"
)

func TestIsScheduleBFridayOff(t *testing.T) {
	tests := []struct {
		name    string
		date    string
		anchor  string
		wantOff bool
	}{
		{
			name:    "Not a Friday - Thursday",
			date:    "2026-04-09",
			anchor:  "2026-04-03",
			wantOff: false,
		},
		{
			name:    "Not a Friday - Monday",
			date:    "2026-04-13",
			anchor:  "2026-04-03",
			wantOff: false,
		},
		{
			name:    "Use real anchor April 3 (Friday off)",
			date:    "2026-04-03",
			anchor:  "2026-04-03",
			wantOff: true,
		},
		{
			name:    "One week after April 3 (work Friday)",
			date:    "2026-04-10",
			anchor:  "2026-04-03",
			wantOff: false,
		},
		{
			name:    "Two weeks after April 3 (off Friday)",
			date:    "2026-04-17",
			anchor:  "2026-04-03",
			wantOff: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			date, _ := time.Parse("2006-01-02", tt.date)
			got := isScheduleBFridayOff(date, tt.anchor)
			if got != tt.wantOff {
				t.Errorf("isScheduleBFridayOff(%s, %s) = %v, want %v",
					tt.date, tt.anchor, got, tt.wantOff)
			}
		})
	}
}

func TestIsHoliday(t *testing.T) {
	cfg := &Config{
		Holidays2025: []string{"2025-12-25", "2025-01-01"},
		Holidays2026: []string{"2026-12-25", "2026-07-04"},
	}

	tests := []struct {
		name    string
		date    string
		wantHol bool
	}{
		{
			name:    "2025 Christmas",
			date:    "2025-12-25",
			wantHol: true,
		},
		{
			name:    "2026 July 4th",
			date:    "2026-07-04",
			wantHol: true,
		},
		{
			name:    "Regular workday",
			date:    "2026-04-15",
			wantHol: false,
		},
		{
			name:    "2025 New Year",
			date:    "2025-01-01",
			wantHol: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isHoliday(tt.date, cfg)
			if got != tt.wantHol {
				t.Errorf("isHoliday(%s) = %v, want %v",
					tt.date, got, tt.wantHol)
			}
		})
	}
}

func TestCalculateWorkingDays(t *testing.T) {
	cfg := &Config{
		ScheduleBanchor: "2026-04-03",
		Holidays2026:    []string{"2026-04-09"},
	}

	tests := []struct {
		name      string
		startDate string
		endDate   string
		wantDays  int
	}{
		{
			name:      "One week Mon-Fri (no holidays, includes Schedule B off Friday)",
			startDate: "2026-04-13",
			endDate:   "2026-04-17",
			wantDays:  4,
		},
		{
			name:      "Week Mon-Thu with holiday Thursday",
			startDate: "2026-04-06",
			endDate:   "2026-04-09",
			wantDays:  3,
		},
		{
			name:      "Full week Mon-Sun (5 work days, Fri is work day)",
			startDate: "2026-04-06",
			endDate:   "2026-04-12",
			wantDays:  4,
		},
		{
			name:      "Single work day (Monday)",
			startDate: "2026-04-13",
			endDate:   "2026-04-13",
			wantDays:  1,
		},
		{
			name:      "Single day (Saturday)",
			startDate: "2026-04-18",
			endDate:   "2026-04-18",
			wantDays:  0,
		},
		{
			name:      "Two weeks (10 work days minus 1 Schedule B Friday)",
			startDate: "2026-04-13",
			endDate:   "2026-04-24",
			wantDays:  9,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateWorkingDays(tt.startDate, tt.endDate, cfg)
			if got != tt.wantDays {
				t.Errorf("CalculateWorkingDays(%s, %s) = %d, want %d",
					tt.startDate, tt.endDate, got, tt.wantDays)
			}
		})
	}
}

func TestCalculateWorkingDaysNoScheduleB(t *testing.T) {
	cfg := &Config{
		ScheduleBanchor: "",
		Holidays2026:    []string{},
	}

	got := CalculateWorkingDays("2026-04-06", "2026-04-12", cfg)
	want := 5

	if got != want {
		t.Errorf("CalculateWorkingDays (no Schedule B) = %d, want %d", got, want)
	}
}
