package task

import (
	"context"
	"fmt"
	"testing"
	"time"

	"example.com/taskservice/internal/repository/mock"
)

func TestRunCleanup(t *testing.T) {
	tests := []struct {
		name         string
		batchResults []int64
		batchError   error
		cfg          CleanupConfig
		wantDeleted  int64
		wantErr      bool
	}{
		{
			name:         "deletes_in_batches",
			batchResults: []int64{1000, 500},
			cfg:          CleanupConfig{OlderThan: 14, BatchSize: 1000},
			wantDeleted:  1500,
		},
		{
			name:         "error_stops_cleanup",
			batchResults: []int64{1000},
			batchError:   fmt.Errorf("db connection lost"),
			cfg:          CleanupConfig{OlderThan: 14, BatchSize: 1000},
			wantDeleted:  1000,
			wantErr:      true,
		},
		{
			name:         "nothing_to_delete",
			batchResults: []int64{0},
			cfg:          CleanupConfig{OlderThan: 14, BatchSize: 1000},
			wantDeleted:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			callCount := 0
			mockRepo := &mock.MockRepository{
				DeleteOldInstancesFunc: func(ctx context.Context, cutoff time.Time, batchSize int) (int64, error) {
					if callCount < len(tt.batchResults) {
						res := tt.batchResults[callCount]
						callCount++
						return res, tt.batchError
					}
					return 0, nil
				},
			}

			svc := NewService(mockRepo)

			got, err := svc.RunCleanup(context.Background(), tt.cfg)

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if got != tt.wantDeleted {
				t.Errorf("deleted = %d, want %d", got, tt.wantDeleted)
			}
		})
	}
}
