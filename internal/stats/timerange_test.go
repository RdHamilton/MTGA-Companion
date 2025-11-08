package stats

import (
	"testing"
	"time"
)

func TestWeekRange(t *testing.T) {
	// Use a fixed time for testing: Wednesday, January 10, 2024
	fixedTime := time.Date(2024, 1, 10, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name       string
		offset     int
		wantStart  time.Time
		wantEnd    time.Time
		wantPeriod string
	}{
		{
			name:       "Current week",
			offset:     0,
			wantStart:  time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC),  // Monday, Jan 8
			wantEnd:    time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC), // Monday, Jan 15
			wantPeriod: "2024-01-08 to 2024-01-14",
		},
		{
			name:       "Last week",
			offset:     -1,
			wantStart:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), // Monday, Jan 1
			wantEnd:    time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC), // Monday, Jan 8
			wantPeriod: "2024-01-01 to 2024-01-07",
		},
		{
			name:       "Two weeks ago",
			offset:     -2,
			wantStart:  time.Date(2023, 12, 25, 0, 0, 0, 0, time.UTC), // Monday, Dec 25
			wantEnd:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),   // Monday, Jan 1
			wantPeriod: "2023-12-25 to 2023-12-31",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := WeekRangeFrom(fixedTime, tt.offset)

			if !tr.Start.Equal(tt.wantStart) {
				t.Errorf("WeekRangeFrom(%v, %d).Start = %v, want %v", fixedTime, tt.offset, tr.Start, tt.wantStart)
			}

			if !tr.End.Equal(tt.wantEnd) {
				t.Errorf("WeekRangeFrom(%v, %d).End = %v, want %v", fixedTime, tt.offset, tr.End, tt.wantEnd)
			}

			if got := tr.FormatPeriod(); got != tt.wantPeriod {
				t.Errorf("WeekRangeFrom(%v, %d).FormatPeriod() = %v, want %v", fixedTime, tt.offset, got, tt.wantPeriod)
			}
		})
	}
}

func TestWeekRangeFromDate(t *testing.T) {
	tests := []struct {
		name      string
		date      time.Time
		wantStart time.Time
		wantEnd   time.Time
	}{
		{
			name:      "Wednesday",
			date:      time.Date(2024, 1, 10, 12, 0, 0, 0, time.UTC), // Wednesday
			wantStart: time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC),   // Monday
			wantEnd:   time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),  // Next Monday
		},
		{
			name:      "Monday",
			date:      time.Date(2024, 1, 8, 12, 0, 0, 0, time.UTC), // Monday
			wantStart: time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC),  // Same Monday
			wantEnd:   time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC), // Next Monday
		},
		{
			name:      "Sunday",
			date:      time.Date(2024, 1, 14, 12, 0, 0, 0, time.UTC), // Sunday
			wantStart: time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC),   // Previous Monday
			wantEnd:   time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),  // Next Monday
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := WeekRangeFromDate(tt.date)

			if !tr.Start.Equal(tt.wantStart) {
				t.Errorf("WeekRangeFromDate(%v).Start = %v, want %v", tt.date, tr.Start, tt.wantStart)
			}

			if !tr.End.Equal(tt.wantEnd) {
				t.Errorf("WeekRangeFromDate(%v).End = %v, want %v", tt.date, tr.End, tt.wantEnd)
			}
		})
	}
}

func TestMonthRange(t *testing.T) {
	// Use a fixed time for testing: January 15, 2024
	fixedTime := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name       string
		offset     int
		wantStart  time.Time
		wantEnd    time.Time
		wantPeriod string
	}{
		{
			name:       "Current month",
			offset:     0,
			wantStart:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			wantEnd:    time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
			wantPeriod: "2024-01-01 to 2024-01-31",
		},
		{
			name:       "Last month",
			offset:     -1,
			wantStart:  time.Date(2023, 12, 1, 0, 0, 0, 0, time.UTC),
			wantEnd:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			wantPeriod: "2023-12-01 to 2023-12-31",
		},
		{
			name:       "Two months ago",
			offset:     -2,
			wantStart:  time.Date(2023, 11, 1, 0, 0, 0, 0, time.UTC),
			wantEnd:    time.Date(2023, 12, 1, 0, 0, 0, 0, time.UTC),
			wantPeriod: "2023-11-01 to 2023-11-30",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := MonthRangeFrom(fixedTime, tt.offset)

			if !tr.Start.Equal(tt.wantStart) {
				t.Errorf("MonthRangeFrom(%v, %d).Start = %v, want %v", fixedTime, tt.offset, tr.Start, tt.wantStart)
			}

			if !tr.End.Equal(tt.wantEnd) {
				t.Errorf("MonthRangeFrom(%v, %d).End = %v, want %v", fixedTime, tt.offset, tr.End, tt.wantEnd)
			}

			if got := tr.FormatPeriod(); got != tt.wantPeriod {
				t.Errorf("MonthRangeFrom(%v, %d).FormatPeriod() = %v, want %v", fixedTime, tt.offset, got, tt.wantPeriod)
			}
		})
	}
}

func TestMonthRangeFromDate(t *testing.T) {
	tests := []struct {
		name      string
		date      time.Time
		wantStart time.Time
		wantEnd   time.Time
		wantName  string
	}{
		{
			name:      "Mid-month",
			date:      time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC),
			wantStart: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			wantEnd:   time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
			wantName:  "January 2024",
		},
		{
			name:      "First day of month",
			date:      time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			wantStart: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			wantEnd:   time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
			wantName:  "January 2024",
		},
		{
			name:      "Last day of month",
			date:      time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC),
			wantStart: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			wantEnd:   time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
			wantName:  "January 2024",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := MonthRangeFromDate(tt.date)

			if !tr.Start.Equal(tt.wantStart) {
				t.Errorf("MonthRangeFromDate(%v).Start = %v, want %v", tt.date, tr.Start, tt.wantStart)
			}

			if !tr.End.Equal(tt.wantEnd) {
				t.Errorf("MonthRangeFromDate(%v).End = %v, want %v", tt.date, tr.End, tt.wantEnd)
			}

			if got := tr.GetMonthName(); got != tt.wantName {
				t.Errorf("MonthRangeFromDate(%v).GetMonthName() = %v, want %v", tt.date, got, tt.wantName)
			}
		})
	}
}

func TestGetWeekLabel(t *testing.T) {
	tests := []struct {
		offset int
		want   string
	}{
		{0, "This Week"},
		{-1, "Last Week"},
		{-2, "Two Weeks Ago"},
		{-5, "5 Weeks Ago"},
		{1, "1 Weeks From Now"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := GetWeekLabel(tt.offset); got != tt.want {
				t.Errorf("GetWeekLabel(%d) = %v, want %v", tt.offset, got, tt.want)
			}
		})
	}
}

func TestGetMonthLabel(t *testing.T) {
	tests := []struct {
		offset int
		want   string
	}{
		{0, "This Month"},
		{-1, "Last Month"},
		{-2, "Two Months Ago"},
		{-5, "5 Months Ago"},
		{1, "1 Months From Now"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := GetMonthLabel(tt.offset); got != tt.want {
				t.Errorf("GetMonthLabel(%d) = %v, want %v", tt.offset, got, tt.want)
			}
		})
	}
}
