package cron

import (
	"context"
	"time"

	usecase "example.com/taskservice/internal/usecase/task"
	"log/slog"
)

const INTERVAL = 1 * time.Minute

type Cron struct {
	service *usecase.Service
}

func New(s *usecase.Service) Job {
	return &Cron{service: s}
}

func (c *Cron) Start(ctx context.Context) {
	slog.Info("cron started", slog.Duration("interval", INTERVAL))

	ticker := time.NewTicker(INTERVAL)
	defer ticker.Stop()

	c.run(ctx)

	for {
		select {
		case <-ctx.Done():
			slog.Info("cron stopped")
			return
		case <-ticker.C:
			c.run(ctx)
		}
	}
}

func (c *Cron) run(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("cron panicked", "recover", r)
		}
	}()

	now := time.Now().UTC()
	to := now.Add(48 * time.Hour)

	created, err := c.service.Materialize(ctx, now, to)
	if err != nil {
		slog.Error(ErrMaterializeFailed.Error(), slog.Any("error", ErrMaterializeFailed))
		return
	}
	slog.Info("cron materialized", slog.Int("count", created))
}
