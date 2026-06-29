package task

import (
	"time"
)

type Status string

const (
	StatusNew        Status = "new"
	StatusInProgress Status = "in_progress"
	StatusDone       Status = "done"
)

type Task struct {
	ID                 int64             `json:"id"`
	Title              string            `json:"title"`
	Description        string            `json:"description"`
	Status             Status            `json:"status"`
	DueDate            *time.Time        `json:"due_date"`
	Tags               []Tag             `json:"tags,omitempty"`
	IsRecurrence       bool              `json:"is_recurrence"`
	RecurrenceParentID *int64            `json:"recurrence_parent_id,omitempty"`
	RecurrenceType     RecurrenceType    `json:"recurrence_type,omitempty"`
	RecurrenceConfig   *RecurrenceConfig `json:"recurrence_config,omitempty"`
	CreatedAt          time.Time         `json:"created_at"`
	UpdatedAt          time.Time         `json:"updated_at"`
}

func (s Status) Valid() bool {
	switch s {
	case StatusNew, StatusInProgress, StatusDone:
		return true
	default:
		return false
	}
}
