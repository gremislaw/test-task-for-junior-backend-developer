package task

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

const (
	OLDERTHAN = 14
	BATCHSIZE = 1000
)

type CleanupConfig struct {
	OlderThan int
	BatchSize int
}

func LoadCleanupConfig() CleanupConfig {
	cfg := CleanupConfig{
		OlderThan: OLDERTHAN,
		BatchSize: BATCHSIZE,
	}
	return cfg
}

func (s *Service) RunCleanup(ctx context.Context, cfg CleanupConfig) (int64, error) {
	cutoff := time.Now().UTC().AddDate(0, 0, -cfg.OlderThan)
	var totalDeleted int64

	slog.Info("cleanup started",
		slog.Time("cutoff", cutoff),
		slog.Int("older_than", cfg.OlderThan),
	)

	for {
		deleted, err := s.repo.DeleteOldInstances(ctx, cutoff, cfg.BatchSize)
		if err != nil {
			return totalDeleted, fmt.Errorf("%w: %w", ErrDeleteOldInstance, err)
		}
		totalDeleted += deleted
		slog.Debug("cleanup batch completed", slog.Int64("deleted", deleted))

		if deleted < int64(cfg.BatchSize) {
			break
		}

		select {
		case <-ctx.Done():
			return totalDeleted, ctx.Err()
		case <-time.After(100 * time.Millisecond):
		}
	}

	if totalDeleted > 0 {
		slog.Info("cleanup finished", slog.Int64("total_deleted", totalDeleted))
	}
	return totalDeleted, nil
}
