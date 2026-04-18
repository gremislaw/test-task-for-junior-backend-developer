package task

import (
	taskdomain "example.com/taskservice/internal/domain/task"
	"testing"
	"time"
)

func TestDebugCalculateDates(t *testing.T) {
	cfg := &taskdomain.RecurrenceConfig{IntervalDays: 1}
	dueDate := time.Date(2026, 5, 1, 10, 0, 0, 0, time.UTC)
	from := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 5, 10, 23, 59, 59, 0, time.UTC)

	dates := CalculateRecurrenceDates(taskdomain.TypeDaily, cfg, dueDate, from, to)

	t.Logf("Generated %d dates:", len(dates))
	for _, d := range dates {
		t.Logf("  - %s", d.Format(time.RFC3339))
	}
}
