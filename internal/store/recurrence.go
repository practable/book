package store

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/practable/book/internal/interval"
)

// WeeklyRecurrence describes a finite, civil-time weekly interval. EndDate is
// inclusive. Occurrences are converted to UTC when the manifest is loaded.
type WeeklyRecurrence struct {
	Timezone   string   `json:"timezone" yaml:"timezone"`
	StartDate  string   `json:"start_date" yaml:"start_date"`
	EndDate    string   `json:"end_date" yaml:"end_date"`
	Weekdays   []string `json:"weekdays" yaml:"weekdays"`
	StartTime  string   `json:"start_time" yaml:"start_time"`
	EndTime    string   `json:"end_time" yaml:"end_time"`
	Exceptions []string `json:"exceptions,omitempty" yaml:"exceptions,omitempty"`
}

func parseCivilDate(value string, location *time.Location) (time.Time, error) {
	return time.ParseInLocation("2006-01-02", value, location)
}

func parseClock(value string) (int, int, error) {
	parsed, err := time.Parse("15:04", value)
	if err != nil {
		return 0, 0, err
	}
	return parsed.Hour(), parsed.Minute(), nil
}

func recurrenceWeekdays(values []string) (map[time.Weekday]bool, error) {
	names := map[string]time.Weekday{"sun": time.Sunday, "sunday": time.Sunday, "mon": time.Monday, "monday": time.Monday, "tue": time.Tuesday, "tuesday": time.Tuesday, "wed": time.Wednesday, "wednesday": time.Wednesday, "thu": time.Thursday, "thursday": time.Thursday, "fri": time.Friday, "friday": time.Friday, "sat": time.Saturday, "saturday": time.Saturday}
	result := make(map[time.Weekday]bool)
	for _, value := range values {
		weekday, ok := names[strings.ToLower(strings.TrimSpace(value))]
		if !ok {
			return nil, fmt.Errorf("unknown weekday %q", value)
		}
		result[weekday] = true
	}
	if len(result) == 0 {
		return nil, errors.New("at least one weekday is required")
	}
	return result, nil
}

func materializeWeekly(rule WeeklyRecurrence) ([]interval.Interval, error) {
	location, err := time.LoadLocation(rule.Timezone)
	if err != nil {
		return nil, fmt.Errorf("invalid timezone: %w", err)
	}
	first, err := parseCivilDate(rule.StartDate, location)
	if err != nil {
		return nil, fmt.Errorf("invalid start_date: %w", err)
	}
	last, err := parseCivilDate(rule.EndDate, location)
	if err != nil {
		return nil, fmt.Errorf("invalid end_date: %w", err)
	}
	if last.Before(first) {
		return nil, errors.New("end_date precedes start_date")
	}
	if last.Sub(first) > 10*366*24*time.Hour {
		return nil, errors.New("recurrence exceeds ten-year materialization limit")
	}
	weekdays, err := recurrenceWeekdays(rule.Weekdays)
	if err != nil {
		return nil, err
	}
	startHour, startMinute, err := parseClock(rule.StartTime)
	if err != nil {
		return nil, fmt.Errorf("invalid start_time: %w", err)
	}
	endHour, endMinute, err := parseClock(rule.EndTime)
	if err != nil {
		return nil, fmt.Errorf("invalid end_time: %w", err)
	}
	exceptions := make(map[string]bool)
	for _, value := range rule.Exceptions {
		if _, err := parseCivilDate(value, location); err != nil {
			return nil, fmt.Errorf("invalid exception %q", value)
		}
		exceptions[value] = true
	}
	result := make([]interval.Interval, 0)
	for day := first; !day.After(last); day = day.AddDate(0, 0, 1) {
		if !weekdays[day.Weekday()] || exceptions[day.Format("2006-01-02")] {
			continue
		}
		start := time.Date(day.Year(), day.Month(), day.Day(), startHour, startMinute, 0, 0, location)
		end := time.Date(day.Year(), day.Month(), day.Day(), endHour, endMinute, 0, 0, location)
		if !end.After(start) {
			end = end.AddDate(0, 0, 1)
		}
		result = append(result, interval.Interval{Start: start.UTC(), End: end.UTC()})
	}
	return result, nil
}

func expandWindow(window Window) (Window, error) {
	expanded := Window{Allowed: append([]interval.Interval(nil), window.Allowed...), Denied: append([]interval.Interval(nil), window.Denied...)}
	for _, rule := range window.RecurringAllowed {
		occurrences, err := materializeWeekly(rule)
		if err != nil {
			return Window{}, err
		}
		expanded.Allowed = append(expanded.Allowed, occurrences...)
	}
	for _, rule := range window.RecurringDenied {
		occurrences, err := materializeWeekly(rule)
		if err != nil {
			return Window{}, err
		}
		expanded.Denied = append(expanded.Denied, occurrences...)
	}
	return expanded, nil
}
