package calendar

import (
	"time"
)

type Config struct {
	ScheduleBanchor string
	Holidays2025    []string
	Holidays2026    []string
}

func CalculateWorkingDays(startDate, endDate string, cfg *Config) int {
	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return 0
	}
	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return 0
	}

	workingDays := 0
	current := start

	for !current.After(end) {
		if current.Weekday() == time.Saturday || current.Weekday() == time.Sunday {
			current = current.AddDate(0, 0, 1)
			continue
		}

		dateStr := current.Format("2006-01-02")
		if isHoliday(dateStr, cfg) {
			current = current.AddDate(0, 0, 1)
			continue
		}

		if cfg.ScheduleBanchor != "" && isScheduleBFridayOff(current, cfg.ScheduleBanchor) {
			current = current.AddDate(0, 0, 1)
			continue
		}

		workingDays++
		current = current.AddDate(0, 0, 1)
	}

	return workingDays
}

func isHoliday(date string, cfg *Config) bool {
	for _, holiday := range cfg.Holidays2025 {
		if date == holiday {
			return true
		}
	}

	for _, holiday := range cfg.Holidays2026 {
		if date == holiday {
			return true
		}
	}

	return false
}

func isScheduleBFridayOff(date time.Time, anchorStr string) bool {
	if date.Weekday() != time.Friday {
		return false
	}

	anchor, err := time.Parse("2006-01-02", anchorStr)
	if err != nil {
		return false
	}

	daysDiff := date.Sub(anchor).Hours() / 24
	weeksDiff := int(daysDiff / 7)

	if weeksDiff < 0 {
		weeksDiff = -weeksDiff - 1
	}

	return weeksDiff%2 == 0
}
