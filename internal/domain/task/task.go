package task

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

type Status string

const (
	StatusNew        Status = "new"
	StatusInProgress Status = "in_progress"
	StatusDone       Status = "done"
)

type RecurrenceType string

const (
	TypeDaily    RecurrenceType = "daily"
	TypeMonthly  RecurrenceType = "monthly"
	TypeSpecific RecurrenceType = "specific"
	TypeEvenOdd  RecurrenceType = "even_odd"
)

type RecurrenceConfig struct {
	IntervalDays int        `json:"interval_days,omitempty"`
	DaysOfMonth  []int      `json:"days_of_month,omitempty"`
	Parity       string     `json:"parity,omitempty"`
	EndDate      *time.Time `json:"end_date,omitempty"`
}

type Task struct {
	ID                 int64             `json:"id"`
	Title              string            `json:"title"`
	Description        string            `json:"description"`
	Status             Status            `json:"status"`
	DueDate            *time.Time        `json:"due_date"`
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

func (rc *RecurrenceConfig) Valid(rt RecurrenceType) bool {
	if rc == nil {
		return false
	}

	switch rt {
	case TypeDaily:
		if rc.IntervalDays <= 0 || rc.IntervalDays > 15 {
			return false
		}
	case TypeMonthly, TypeSpecific:
		if len(rc.DaysOfMonth) == 0 {
			return false
		}
		for _, d := range rc.DaysOfMonth {
			if d < 1 || d > 28 {
				return false
			}
		}
	case TypeEvenOdd:
		if rc.Parity != "even" && rc.Parity != "odd" {
			return false
		}
	default:
		return false
	}
	return true

	// TODO: залогировать
}

func (c *RecurrenceConfig) Scan(value any) error {
	if value == nil {
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("expected []byte or string for JSONB, got %T", value)
	}

	return json.Unmarshal(bytes, c)
}

func (c *RecurrenceConfig) Value() (driver.Value, error) {
	if c == nil {
		return nil, nil
	}
	return json.Marshal(c)
}
