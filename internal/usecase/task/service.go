package task

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type Service struct {
	repo Repository
	now  func() time.Time
}

func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
		now:  func() time.Time { return time.Now().UTC() },
	}
}

func (s *Service) SetNow(fn func() time.Time) {
	s.now = fn
}

func (s *Service) Create(ctx context.Context, input CreateInput) (*taskdomain.Task, error) {
	normalized, err := validateCreateInput(input)
	if err != nil {
		return nil, err
	}

	model := &taskdomain.Task{
		Title:            normalized.Title,
		Description:      normalized.Description,
		Status:           normalized.Status,
		DueDate:          normalized.DueDate,
		IsRecurrence:     normalized.IsRecurrence,
		RecurrenceType:   normalized.RecurrenceType,
		RecurrenceConfig: normalized.RecurrenceConfig,
	}
	now := s.now()
	model.CreatedAt = now
	model.UpdatedAt = now

	created, err := s.repo.Create(ctx, model)
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (s *Service) GetByID(ctx context.Context, id int64) (*taskdomain.Task, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	return s.repo.GetByID(ctx, id)
}

func (s *Service) Update(ctx context.Context, id int64, input UpdateInput) (*taskdomain.Task, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	normalized, err := validateUpdateInput(input)
	if err != nil {
		return nil, err
	}

	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if existing.RecurrenceParentID != nil {
		if input.IsRecurrence {
			return nil, ErrInstanceCannotBeTemplate
		}
		if input.RecurrenceType != "" || input.RecurrenceConfig != nil {
			return nil, ErrInstanceFieldsImmutable
		}
	}

	if normalized.Title != "" {
		existing.Title = normalized.Title
	}
	if normalized.Description != "" {
		existing.Description = normalized.Description
	}
	if normalized.Status != "" {
		existing.Status = normalized.Status
	}
	if normalized.DueDate != nil {
		existing.DueDate = normalized.DueDate
	}

	if normalized.RecurrenceType != "" {
		existing.RecurrenceType = normalized.RecurrenceType
	}
	if normalized.RecurrenceConfig != nil {
		existing.RecurrenceConfig = normalized.RecurrenceConfig
	}
	if normalized.IsRecurrence {
		existing.IsRecurrence = normalized.IsRecurrence
	}

	existing.UpdatedAt = s.now()

	updated, err := s.repo.Update(ctx, existing)
	if err != nil {
		return nil, err
	}

	return updated, nil
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	return s.repo.Delete(ctx, id)
}

func (s *Service) List(ctx context.Context, from, to time.Time, cursorDate *time.Time, cursorID *int64, limit int) ([]*taskdomain.Task, string, error) {
	created, err := s.Materialize(ctx, from, to)
	if err != nil {
		slog.Warn("background materialize failed, existing tasks will be served",
			slog.Time("from", from),
			slog.Time("to", to),
			slog.String("error", err.Error()),
		)
	}
	slog.Info("lazy materialized", slog.Int("count", created))

	tasks, hasMore, err := s.repo.ListByRange(ctx, from, to, cursorDate, cursorID, limit)
	if err != nil {
		return nil, "", fmt.Errorf("%w: cannot list tasks by range: %w", ErrMaterialize, err)
	}

	var nextCursor string
	if hasMore && len(tasks) > 0 {
		last := tasks[len(tasks)-1]
		nextCursor = fmt.Sprintf("%s_%d", last.DueDate.Format(time.RFC3339), last.ID)
	}

	return tasks, nextCursor, nil
}

func (s *Service) DetachInstance(ctx context.Context, instanceID int64) (*taskdomain.Task, error) {
	inst, err := s.repo.GetByID(ctx, instanceID)
	if err != nil { return nil, err }
	if inst.RecurrenceParentID == nil {
		return nil, errs.ErrCannotDetachNonInstance
	}

	parent, err := s.repo.GetByID(ctx, *inst.RecurrenceParentID)
	if err != nil { return nil, err }

	if parent.RecurrenceConfig == nil {
		parent.RecurrenceConfig = &taskdomain.RecurrenceConfig{}
	}
	parent.RecurrenceConfig.ExcludedDates = append(
		parent.RecurrenceConfig.ExcludedDates, *inst.DueDate,
	)
	if err := s.repo.Update(ctx, parent); err != nil { return nil, err }

	inst.RecurrenceParentID = nil
	inst.IsRecurrence = false
	inst.RecurrenceType = ""
	inst.RecurrenceConfig = nil
	inst.UpdatedAt = s.now()

	return s.repo.Update(ctx, inst)
}

func validateCreateInput(input CreateInput) (CreateInput, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)

	if input.Title == "" {
		return CreateInput{}, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}

	if input.Status == "" {
		input.Status = taskdomain.StatusNew
	}

	if !input.Status.Valid() {
		return CreateInput{}, fmt.Errorf("%w: invalid status", ErrInvalidInput)
	}

	if input.DueDate == nil {
		defaultDue := time.Now()
		input.DueDate = &defaultDue
	}

	if input.IsRecurrence {
		if input.RecurrenceType == "" || input.RecurrenceConfig == nil {
			return CreateInput{}, fmt.Errorf("%w: recurrence_type and recurrence_config are required for recurrence", ErrInvalidInput)
		}
		if !input.RecurrenceConfig.Valid(input.RecurrenceType) {
			return CreateInput{}, fmt.Errorf("%w: invalid recurrence config", ErrInvalidInput)
		}
	}

	return input, nil
}

func validateUpdateInput(input UpdateInput) (UpdateInput, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)

	if input.Title == "" {
		return UpdateInput{}, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}

	if !input.Status.Valid() {
		return UpdateInput{}, fmt.Errorf("%w: invalid status", ErrInvalidInput)
	}

	if input.DueDate == nil {
		defaultDue := time.Now()
		input.DueDate = &defaultDue
	}

	if input.IsRecurrence {
		if input.RecurrenceType == "" || input.RecurrenceConfig == nil {
			return UpdateInput{}, fmt.Errorf("%w: recurrence_type and recurrence_config are required for recurrence", ErrInvalidInput)
		}
		if !input.RecurrenceConfig.Valid(input.RecurrenceType) {
			return UpdateInput{}, fmt.Errorf("%w: invalid recurrence config", ErrInvalidInput)
		}
	}

	return input, nil
}
