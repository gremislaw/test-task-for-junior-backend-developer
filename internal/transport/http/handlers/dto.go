package handlers

import (
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type taskMutationDTO struct {
	Title            string                       `json:"title"`
	Description      string                       `json:"description"`
	Status           taskdomain.Status            `json:"status"`
	DueDate          *time.Time                   `json:"due_date"`
	IsRecurrence     bool                         `json:"is_recurrence,omitempty"`
	RecurrenceType   taskdomain.RecurrenceType    `json:"recurrence_type,omitempty"`
	RecurrenceConfig *taskdomain.RecurrenceConfig `json:"recurrence_config,omitempty"`
}

type taskDTO struct {
	ID                 int64                        `json:"id"`
	Title              string                       `json:"title"`
	Description        string                       `json:"description"`
	Status             taskdomain.Status            `json:"status"`
	DueDate            *time.Time                   `json:"due_date"`
	CreatedAt          time.Time                    `json:"created_at"`
	UpdatedAt          time.Time                    `json:"updated_at"`
	IsRecurrence       bool                         `json:"is_recurrence"`
	RecurrenceParentID *int64                       `json:"recurrence_parent_id,omitempty"`
	RecurrenceType     taskdomain.RecurrenceType    `json:"recurrence_type,omitempty"`
	RecurrenceConfig   *taskdomain.RecurrenceConfig `json:"recurrence_config,omitempty"`
}

type taskListDTO struct {
	Tasks  []taskDTO `json:"tasks"`
	Cursor string    `json:"cursor,omitempty"`
	Limit  int       `json:"limit"`
}

func newTaskDTO(task *taskdomain.Task) taskDTO {
	return taskDTO{
		ID:                 task.ID,
		Title:              task.Title,
		Description:        task.Description,
		Status:             task.Status,
		DueDate:            task.DueDate,
		CreatedAt:          task.CreatedAt,
		UpdatedAt:          task.UpdatedAt,
		IsRecurrence:       task.IsRecurrence,
		RecurrenceParentID: task.RecurrenceParentID,
		RecurrenceType:     task.RecurrenceType,
		RecurrenceConfig:   task.RecurrenceConfig,
	}
}
