package cron

import (
	"context"
	"time"

	"log/slog"
	usecase "example.com/taskservice/internal/usecase/task"
)

type Cron struct {
	uc usecase.Service
}

func New(uc usecase.Service) *Cron {
	return &Cron{uc: uc}
}

func (c *Cron) Start(ctx context.Context) {
	slog.Info("cron started", slog.Duration("interval", 10*time.Minute))
	
	ticker := time.NewTicker(10 * time.Minute)
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
	now := time.Now().UTC()
	to := now.Add(24 * time.Hour)

	created, err := c.uc.Materialize(ctx, now, to)
	if err != nil {
		slog.Error(ErrMaterializeFailed.Error(), slog.Any("error", ErrMaterializeFailed))
		return
	}
	if created > 0 {
		slog.Info("cron materialized", slog.Int("count", created))
	}
}