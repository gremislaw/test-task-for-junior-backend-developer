package task

import (
	"context"
	"fmt"
	"testing"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
	"example.com/taskservice/internal/repository/mock"
)

func TestMaterialize(t *testing.T) {
	utc := time.UTC
	fixedNow := time.Date(2026, 4, 18, 12, 0, 0, 0, utc)

	tests := []struct {
		name           string
		parents        []taskdomain.Task
		existingDates  map[int64][]time.Time
		batchCreateErr error
		wantCreated    int
		wantErr        bool
	}{
		{
			name: "creates_new_instances",
			parents: []taskdomain.Task{
				{
					ID:               1,
					Title:            "Test",
					DueDate:          ptrTime(time.Date(2026, 4, 18, 10, 0, 0, 0, utc)),
					IsRecurrence:     true,
					RecurrenceType:   "daily",
					RecurrenceConfig: &taskdomain.RecurrenceConfig{IntervalDays: 1},
				},
			},
			existingDates: map[int64][]time.Time{},
			wantCreated:   1, // 2026-04-19 в диапазоне [18, 20]
		},
		{
			name: "skips_existing_instances",
			parents: []taskdomain.Task{
				{
					ID:               1,
					Title:            "Test",
					DueDate:          ptrTime(time.Date(2026, 4, 18, 10, 0, 0, 0, utc)),
					IsRecurrence:     true,
					RecurrenceType:   "daily",
					RecurrenceConfig: &taskdomain.RecurrenceConfig{IntervalDays: 1},
				},
			},
			existingDates: map[int64][]time.Time{
				1: {time.Date(2026, 4, 19, 10, 0, 0, 0, utc)},
			},
			wantCreated: 0,
		},
		{
			name: "batch_create_error_continue",
			parents: []taskdomain.Task{
				{ID: 1, Title: "A", DueDate: ptrTime(time.Date(2026, 4, 18, 10, 0, 0, 0, utc)), IsRecurrence: true, RecurrenceType: "daily", RecurrenceConfig: &taskdomain.RecurrenceConfig{IntervalDays: 1}},
				{ID: 2, Title: "B", DueDate: ptrTime(time.Date(2026, 4, 18, 10, 0, 0, 0, utc)), IsRecurrence: true, RecurrenceType: "daily", RecurrenceConfig: &taskdomain.RecurrenceConfig{IntervalDays: 1}},
			},
			existingDates:  map[int64][]time.Time{},
			batchCreateErr: fmt.Errorf("db error"),
			wantCreated:    0,
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mock.MockRepository{
				ListRecurrenceParentsFunc: func(ctx context.Context, from, to time.Time) ([]taskdomain.Task, error) {
					return tt.parents, nil
				},
				GetRecurrenceInstanceDatesFunc: func(ctx context.Context, parentID int64, from, to time.Time) ([]time.Time, error) {
					return tt.existingDates[parentID], nil
				},
				BatchCreateFunc: func(ctx context.Context, tasks []*taskdomain.Task) error {
					return tt.batchCreateErr
				},
			}

			svc := NewService(mockRepo)
			svc.SetNow(func() time.Time { return fixedNow })

			from := time.Date(2026, 4, 18, 0, 0, 0, 0, utc)
			to := time.Date(2026, 4, 20, 0, 0, 0, 0, utc)

			got, err := svc.Materialize(context.Background(), from, to)

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if got != tt.wantCreated {
				t.Errorf("created = %d, want %d", got, tt.wantCreated)
			}
		})
	}
}

func ptrTime(t time.Time) *time.Time {
	return &t
}
