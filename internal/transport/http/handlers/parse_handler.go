package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
	"example.com/taskservice/internal/llm"
	taskusecase "example.com/taskservice/internal/usecase/task"
)

type ParseHandler struct {
	llmParser   *llm.Parser
	taskService *taskusecase.Service
}

func NewParseHandler(lp *llm.Parser, ts *taskusecase.Service) *ParseHandler {
	return &ParseHandler{llmParser: lp, taskService: ts}
}

func (h *ParseHandler) Serve(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	ext, err := h.extractTask(r.Context(), req.Text)
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, err)
		return
	}

	task := taskusecase.CreateInput{
		Title:       ext.Title,
		Description: ext.Description,
		Status:      taskdomain.StatusNew,
	}

	if ext.DueDate != "" {
		if dt, err := time.Parse(time.RFC3339, ext.DueDate); err == nil {
			task.DueDate = &dt
		} else {
			dt := time.Now().UTC().Add(time.Hour)
			task.DueDate = &dt
		}
	} else {
		dt := time.Now().UTC().Add(time.Hour)
		task.DueDate = &dt
	}

	if ext.RecurrenceType != "none" && ext.RecurrenceType != "" {
		task.IsRecurrence = true
		task.RecurrenceType = taskdomain.RecurrenceType(ext.RecurrenceType)
		cfg := &taskdomain.RecurrenceConfig{}
		if ext.IntervalDays != nil {
			cfg.IntervalDays = *ext.IntervalDays
		}
		cfg.DaysOfMonth = ext.DaysOfMonth
		cfg.Parity = ext.Parity
		task.RecurrenceConfig = cfg
	}

	created, err := h.taskService.Create(r.Context(), task)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusCreated, created)
}

func (h *ParseHandler) extractTask(ctx context.Context, text string) (*llm.TaskExtraction, error) {
	ext, err := h.llmParser.Parse(ctx, text)
	if err != nil {
		return nil, err
	}

	return ext, nil
}
