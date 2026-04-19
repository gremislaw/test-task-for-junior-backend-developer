package task

import (
	"context"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type Repository interface {
	Create(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error)
	GetByID(ctx context.Context, id int64) (*taskdomain.Task, error)
	Update(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context) ([]taskdomain.Task, error)
	ListByRange(ctx context.Context, from, to time.Time, cursorDate *time.Time, cursorID *int64, limit int) ([]*taskdomain.Task, bool, error)
	ListRecurrenceParents(ctx context.Context, from, to time.Time) ([]taskdomain.Task, error)
	GetRecurrenceInstanceDates(ctx context.Context, parentID int64, from, to time.Time) ([]time.Time, error)
	BatchCreate(ctx context.Context, tasks []*taskdomain.Task) error
	DeleteOldInstances(ctx context.Context, olderThan time.Time, limit int) (int64, error)
}

type Usecase interface {
	Create(ctx context.Context, input CreateInput) (*taskdomain.Task, error)
	GetByID(ctx context.Context, id int64) (*taskdomain.Task, error)
	Update(ctx context.Context, id int64, input UpdateInput) (*taskdomain.Task, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, from, to time.Time, cursorDate *time.Time, cursorID *int64, limit int) ([]*taskdomain.Task, string, error)
}

type CreateInput struct {
	Title            string
	Description      string
	Status           taskdomain.Status
	DueDate          *time.Time
	IsRecurrence     bool
	RecurrenceType   taskdomain.RecurrenceType
	RecurrenceConfig *taskdomain.RecurrenceConfig
}

type UpdateInput struct {
	Title            string
	Description      string
	Status           taskdomain.Status
	DueDate          *time.Time
	IsRecurrence     bool
	RecurrenceType   taskdomain.RecurrenceType
	RecurrenceConfig *taskdomain.RecurrenceConfig
}
