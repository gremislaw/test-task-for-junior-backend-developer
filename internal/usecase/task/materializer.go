package task

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
	"example.com/taskservice/internal/observability"
)

func (s *Service) Materialize(ctx context.Context, from, to time.Time) (int, error) {
	if to.Sub(from) > 365*24*time.Hour {
		return 0, ErrRangeTooLarge
	}

	parents, err := s.repo.ListRecurrenceParents(ctx, from, to)
	if err != nil {
		return 0, fmt.Errorf("%w: fetch parents: %w", ErrMaterialize, err)
	}

	total := 0

	for _, parent := range parents {
		dates := CalculateRecurrenceDates(
			parent.RecurrenceType,
			parent.RecurrenceConfig,
			*parent.DueDate,
			from, to,
		)
		if len(dates) == 0 {
			continue
		}

		existing, err := s.repo.GetRecurrenceInstanceDates(ctx, parent.ID, from, to)
		if err != nil {
			slog.Error("fetch existing instances failed", slog.Int64("id", parent.ID), slog.Any("error", err))
			continue
		}

		// дедупликация
		existingSet := make(map[string]struct{}, len(existing))
		for _, d := range existing {
			existingSet[d.Format(time.RFC3339)] = struct{}{}
		}

		var tasks []*taskdomain.Task
		for _, d := range dates {
			if _, exists := existingSet[d.Format(time.RFC3339)]; !exists {
				dueDate := d
				now := s.now()
				instance := taskdomain.Task{
					Title:              parent.Title,
					Description:        parent.Description,
					Status:             taskdomain.StatusNew,
					DueDate:            &dueDate,
					IsRecurrence:       false,
					RecurrenceParentID: &parent.ID,
					RecurrenceType:     "",
					RecurrenceConfig:   nil,
					CreatedAt:          now,
					UpdatedAt:          now,
				}
				tasks = append(tasks, &instance)
			}
		}

		if len(tasks) > 0 {
			if err := s.repo.BatchCreate(ctx, tasks); err != nil {
				slog.Error("fetch existing instances failed", slog.Int64("id", parent.ID), slog.Any("error", err))
				continue
			}
			total += len(tasks)
		}
	}

	observability.CronMaterializeTotal.Add(float64(total))

	return total, nil
}
