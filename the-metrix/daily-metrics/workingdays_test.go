package main

import (
	"testing"
)

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
			got := calculateWorkingDays(tt.startDate, tt.endDate, cfg)
			if got != tt.wantDays {
				t.Errorf("calculateWorkingDays(%s, %s) = %d, want %d",
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

	got := calculateWorkingDays("2026-04-06", "2026-04-12", cfg)
	want := 5

	if got != want {
		t.Errorf("calculateWorkingDays (no Schedule B) = %d, want %d", got, want)
	}
}
