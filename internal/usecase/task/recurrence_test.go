package task

import (
	"testing"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

func TestCalculateDates(t *testing.T) {
	utc := time.UTC
	dueBase := time.Date(2026, 4, 18, 10, 0, 0, 0, utc)

	tests := []struct {
		name      string
		recType   string
		cfg       *taskdomain.RecurrenceConfig
		from      time.Time
		to        time.Time
		wantCount int
		wantFirst string
	}{
		{name: "daily_1d_basic", recType: "daily", cfg: &taskdomain.RecurrenceConfig{IntervalDays: 1},
			from: time.Date(2026, 4, 19, 0, 0, 0, 0, utc), to: time.Date(2026, 4, 21, 0, 0, 0, 0, utc),
			wantCount: 2, wantFirst: "2026-04-19T10:00:00Z"},
		{name: "daily_3d_step", recType: "daily", cfg: &taskdomain.RecurrenceConfig{IntervalDays: 3},
			from: time.Date(2026, 4, 19, 0, 0, 0, 0, utc), to: time.Date(2026, 4, 25, 0, 0, 0, 0, utc),
			wantCount: 2, wantFirst: "2026-04-21T10:00:00Z"},
		{name: "daily_empty_range", recType: "daily", cfg: &taskdomain.RecurrenceConfig{IntervalDays: 1},
			from: time.Date(2026, 5, 1, 0, 0, 0, 0, utc), to: time.Date(2026, 4, 30, 0, 0, 0, 0, utc),
			wantCount: 0},

		{name: "monthly_15th", recType: "monthly", cfg: &taskdomain.RecurrenceConfig{DaysOfMonth: []int{15}},
			from: time.Date(2026, 4, 10, 0, 0, 0, 0, utc), to: time.Date(2026, 6, 10, 0, 0, 0, 0, utc),
			wantCount: 2, wantFirst: "2026-04-15T10:00:00Z"},
		{name: "monthly_nil_config", recType: "monthly", cfg: nil,
			from: time.Date(2026, 4, 1, 0, 0, 0, 0, utc), to: time.Date(2026, 5, 1, 0, 0, 0, 0, utc),
			wantCount: 0},

		{name: "yearly_two_dates", recType: "yearly", 
			cfg: &taskdomain.RecurrenceConfig{DatesOfYear: []string{"01-15", "05-20"}},
			from: time.Date(2026, 1, 1, 0, 0, 0, 0, utc), to: time.Date(2027, 12, 31, 0, 0, 0, 0, utc),
			wantCount: 4, wantFirst: "2026-01-15T10:00:00Z"},

		{name: "yearly_feb29_skip_non_leap", recType: "yearly", 
			cfg: &taskdomain.RecurrenceConfig{DatesOfYear: []string{"02-29"}},
			from: time.Date(2026, 1, 1, 0, 0, 0, 0, utc), to: time.Date(2028, 12, 31, 0, 0, 0, 0, utc),
			wantCount: 1},
		{name: "yearly_invalid_month", recType: "yearly", cfg: &taskdomain.RecurrenceConfig{DatesOfYear: []string{"13-01"}},
			from: time.Date(2026, 1, 1, 0, 0, 0, 0, utc), to: time.Date(2027, 1, 1, 0, 0, 0, 0, utc),
			wantCount: 0},

		{name: "even_odd_odd_days", recType: "even_odd", cfg: &taskdomain.RecurrenceConfig{Parity: "odd"},
			from: time.Date(2026, 4, 1, 0, 0, 0, 0, utc), to: time.Date(2026, 4, 6, 0, 0, 0, 0, utc),
			wantCount: 3, wantFirst: "2026-04-01T10:00:00Z"},
		{name: "even_odd_even_days", recType: "even_odd", cfg: &taskdomain.RecurrenceConfig{Parity: "even"},
			from: time.Date(2026, 4, 1, 0, 0, 0, 0, utc), to: time.Date(2026, 4, 5, 0, 0, 0, 0, utc),
			wantCount: 2, wantFirst: "2026-04-02T10:00:00Z"},
		{name: "even_odd_unknown_parity", recType: "even_odd", cfg: &taskdomain.RecurrenceConfig{Parity: "monthly"},
			from: time.Date(2026, 4, 1, 0, 0, 0, 0, utc), to: time.Date(2026, 4, 5, 0, 0, 0, 0, utc),
			wantCount: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dates := CalculateRecurrenceDates(taskdomain.RecurrenceType(tt.recType), tt.cfg, dueBase, tt.from, tt.to)

			if len(dates) != tt.wantCount {
				t.Errorf("count = %d, want %d", len(dates), tt.wantCount)
				return
			}
			if tt.wantCount > 0 && tt.wantFirst != "" {
				got := dates[0].Format(time.RFC3339)
				if got != tt.wantFirst {
					t.Errorf("first = %v, want %v", got, tt.wantFirst)
				}
			}
		})
	}
}