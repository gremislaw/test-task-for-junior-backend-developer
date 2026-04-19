package task

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"example.com/taskservice/internal/observability"
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
	if cfg.OlderThan < 0 || cfg.BatchSize <= 0 {
		return 0, fmt.Errorf("invalid cleanup config: older_than=%d, batch_size=%d", cfg.OlderThan, cfg.BatchSize)
	}

	cutoff := time.Now().UTC().AddDate(0, 0, -cfg.OlderThan)
	var totalDeleted int64

	slog.Info("cleanup started",
		slog.Time("cutoff", cutoff),
		slog.Int("older_than", cfg.OlderThan),
	)

	for {
		select {
		case <-ctx.Done():
			return totalDeleted, ctx.Err()
		default:
		}

		deleted, err := s.repo.DeleteOldInstances(ctx, cutoff, cfg.BatchSize)
		totalDeleted += deleted
		if err != nil {
			return totalDeleted, fmt.Errorf("%w: %s", ErrDeleteOldInstance, err.Error())
		}

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

	slog.Info("cleanup finished", slog.Int64("total_deleted", totalDeleted))
	observability.CleanupDeletedTotal.Add(float64(totalDeleted))

	return totalDeleted, nil

}
