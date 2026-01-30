package repository

import (
	"context"

	"github.com/beatyman/scan-miners/internal/domain/model"
)

type MinerStatsRepository interface {
	Save(ctx context.Context, stats *model.MinerStats) error
	FindLatestByWorkerID(ctx context.Context, workerID string) (*model.MinerStats, error)
}
