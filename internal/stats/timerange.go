package stats

import (
	"fmt"
	"time"
)

// TimeRange represents a start and end time period.
type TimeRange struct {
	Start time.Time
	End   time.Time
}

// WeekRange calculates the start and end of a week with an offset.
// offset = 0 means current week, -1 means last week, -2 means two weeks ago, etc.
// The week starts on Monday and ends on Sunday.
func WeekRange(offset int) TimeRange {
	return WeekRangeFrom(time.Now(), offset)
}

// WeekRangeFrom calculates the start and end of a week with an offset from a reference time.
// offset = 0 means the week containing referenceTime, -1 means previous week, etc.
// The week starts on Monday and ends on Sunday.
func WeekRangeFrom(referenceTime time.Time, offset int) TimeRange {
	// Get start of current week (Monday)
	weekday := int(referenceTime.Weekday())
	if weekday == 0 {
		weekday = 7 // Sunday is 7 (ISO 8601)
	}
	currentWeekStart := referenceTime.AddDate(0, 0, -weekday+1).Truncate(24 * time.Hour)

	// Apply offset
	weekStart := currentWeekStart.AddDate(0, 0, offset*7)
	weekEnd := weekStart.AddDate(0, 0, 7)

	return TimeRange{
		Start: weekStart,
		End:   weekEnd,
	}
}

// WeekRangeFromDate calculates the week range for a specific date.
// The week containing the given date is calculated (Monday-Sunday).
func WeekRangeFromDate(date time.Time) TimeRange {
	// Get start of week (Monday) for the given date
	weekday := int(date.Weekday())
	if weekday == 0 {
		weekday = 7 // Sunday is 7
	}
	weekStart := date.AddDate(0, 0, -weekday+1).Truncate(24 * time.Hour)
	weekEnd := weekStart.AddDate(0, 0, 7)

	return TimeRange{
		Start: weekStart,
		End:   weekEnd,
	}
}

// MonthRange calculates the start and end of a month with an offset.
// offset = 0 means current month, -1 means last month, -2 means two months ago, etc.
func MonthRange(offset int) TimeRange {
	return MonthRangeFrom(time.Now(), offset)
}

// MonthRangeFrom calculates the start and end of a month with an offset from a reference time.
// offset = 0 means the month containing referenceTime, -1 means previous month, etc.
func MonthRangeFrom(referenceTime time.Time, offset int) TimeRange {
	// Get start of current month
	currentMonthStart := time.Date(referenceTime.Year(), referenceTime.Month(), 1, 0, 0, 0, 0, referenceTime.Location())

	// Apply offset
	monthStart := currentMonthStart.AddDate(0, offset, 0)
	monthEnd := monthStart.AddDate(0, 1, 0)

	return TimeRange{
		Start: monthStart,
		End:   monthEnd,
	}
}

// MonthRangeFromDate calculates the month range for a specific date.
// Returns the full month containing the given date.
func MonthRangeFromDate(date time.Time) TimeRange {
	monthStart := time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, date.Location())
	monthEnd := monthStart.AddDate(0, 1, 0)

	return TimeRange{
		Start: monthStart,
		End:   monthEnd,
	}
}

// FormatPeriod returns a human-readable description of the time period.
func (tr TimeRange) FormatPeriod() string {
	start := tr.Start.Format("2006-01-02")
	end := tr.End.AddDate(0, 0, -1).Format("2006-01-02") // End is exclusive, so subtract 1 day for display
	return fmt.Sprintf("%s to %s", start, end)
}

// GetWeekLabel returns a descriptive label for a week offset.
func GetWeekLabel(offset int) string {
	switch offset {
	case 0:
		return "This Week"
	case -1:
		return "Last Week"
	case -2:
		return "Two Weeks Ago"
	default:
		if offset < 0 {
			return fmt.Sprintf("%d Weeks Ago", -offset)
		}
		return fmt.Sprintf("%d Weeks From Now", offset)
	}
}

// GetMonthLabel returns a descriptive label for a month offset.
func GetMonthLabel(offset int) string {
	switch offset {
	case 0:
		return "This Month"
	case -1:
		return "Last Month"
	case -2:
		return "Two Months Ago"
	default:
		if offset < 0 {
			return fmt.Sprintf("%d Months Ago", -offset)
		}
		return fmt.Sprintf("%d Months From Now", offset)
	}
}

// GetMonthName returns the month name and year for a time range.
func (tr TimeRange) GetMonthName() string {
	return tr.Start.Format("January 2006")
}

// GetWeekDescription returns a descriptive string for a week time range.
func (tr TimeRange) GetWeekDescription() string {
	year, week := tr.Start.ISOWeek()
	return fmt.Sprintf("Week %d of %d", week, year)
}
