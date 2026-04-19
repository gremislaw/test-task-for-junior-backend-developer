package mock

import (
	"context"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type MockRepository struct {
	CreateFunc                        func(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error)
	GetByIDFunc                       func(ctx context.Context, id int64) (*taskdomain.Task, error)
	UpdateFunc                        func(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error)
	DeleteFunc                        func(ctx context.Context, id int64) error
	ListFunc                          func(ctx context.Context) ([]taskdomain.Task, error)
	ListByRangeFunc                   func(ctx context.Context, from, to time.Time, cursorDate *time.Time, cursorID *int64, limit int) ([]*taskdomain.Task, bool, error)
	ListRecurrenceParentsFunc         func(ctx context.Context, from, to time.Time) ([]taskdomain.Task, error)
	GetRecurrenceInstanceDatesFunc    func(ctx context.Context, parentID int64, from, to time.Time) ([]time.Time, error)
	BatchCreateFunc                   func(ctx context.Context, tasks []*taskdomain.Task) error
	DeleteOldInstancesFunc            func(ctx context.Context, olderThan time.Time, limit int) (int64, error)
}

func (m *MockRepository) Create(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, task)
	}
	if task.ID == 0 {
		task.ID = 1
	}
	return task, nil
}

func (m *MockRepository) GetByID(ctx context.Context, id int64) (*taskdomain.Task, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockRepository) Update(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error) {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, task)
	}
	return task, nil
}

func (m *MockRepository) Delete(ctx context.Context, id int64) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

func (m *MockRepository) List(ctx context.Context) ([]taskdomain.Task, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx)
	}
	return nil, nil
}

func (m *MockRepository) ListByRange(ctx context.Context, from, to time.Time, cursorDate *time.Time, cursorID *int64, limit int) ([]*taskdomain.Task, bool, error) {
	if m.ListByRangeFunc != nil {
		return m.ListByRangeFunc(ctx, from, to, cursorDate, cursorID, limit)
	}
	return nil, false, nil
}

func (m *MockRepository) ListRecurrenceParents(ctx context.Context, from, to time.Time) ([]taskdomain.Task, error) {
	if m.ListRecurrenceParentsFunc != nil {
		return m.ListRecurrenceParentsFunc(ctx, from, to)
	}
	return nil, nil
}

func (m *MockRepository) GetRecurrenceInstanceDates(ctx context.Context, parentID int64, from, to time.Time) ([]time.Time, error) {
	if m.GetRecurrenceInstanceDatesFunc != nil {
		return m.GetRecurrenceInstanceDatesFunc(ctx, parentID, from, to)
	}
	return nil, nil
}

func (m *MockRepository) BatchCreate(ctx context.Context, tasks []*taskdomain.Task) error {
	if m.BatchCreateFunc != nil {
		return m.BatchCreateFunc(ctx, tasks)
	}
	return nil
}

func (m *MockRepository) DeleteOldInstances(ctx context.Context, olderThan time.Time, limit int) (int64, error) {
	if m.DeleteOldInstancesFunc != nil {
		return m.DeleteOldInstancesFunc(ctx, olderThan, limit)
	}
	return 0, nil
}