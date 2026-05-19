package main

import (
	"github.com/themetrix/metrics-api/pkg/calendar"
)

func calculateWorkingDays(startDate, endDate string, cfg *Config) int {
	calCfg := &calendar.Config{
		ScheduleBanchor: cfg.ScheduleBanchor,
		Holidays2025:    cfg.Holidays2025,
		Holidays2026:    cfg.Holidays2026,
	}
	return calendar.CalculateWorkingDays(startDate, endDate, calCfg)
}
