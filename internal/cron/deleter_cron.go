package cron

import (
	"context"
	"log/slog"
	"sync"
	"time"

	taskusecase "example.com/taskservice/internal/usecase/task"
)

type CleanupRunner struct {
	uc       taskusecase.Service
	interval time.Duration
	mu       sync.Mutex
}

func NewCleanupRunner(uc taskusecase.Service) *CleanupRunner {
	return &CleanupRunner{uc: uc}
}

func (c *CleanupRunner) Start(ctx context.Context) {
	slog.Info("cleanup started")
	for {
		now := time.Now().UTC()
		next := time.Date(now.Year(), now.Month(), now.Day(), 6, 0, 0, 0, time.UTC)
		if now.After(next) {
			next = next.Add(24 * time.Hour)
		}
		timer := time.NewTimer(next.Sub(now))
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
			c.run(ctx)
		}
	}
}

func (c *CleanupRunner) run(ctx context.Context) {
	if !c.mu.TryLock() {
		slog.Warn("cleanup skipped: previous run still in progress")
		return
	}
	defer c.mu.Unlock()

	defer func() {
		if r := recover(); r != nil {
			slog.Error("cleanup runner panicked", "recover", r)
		}
	}()

	cfg := taskusecase.LoadCleanupConfig()
	deleted, err := c.uc.RunCleanup(ctx, cfg)
	if err != nil {
		slog.Error("cleanup failed", slog.Any("error", err))
		return
	}
	slog.Info("cleanup completed", slog.Int64("deleted", deleted))
}
