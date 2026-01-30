package mysql

import (
	"context"

	"github.com/beatyman/scan-miners/internal/domain/model"
	"github.com/beatyman/scan-miners/internal/domain/repository"
	"gorm.io/gorm"
)

type minerStatsRepository struct {
	db *gorm.DB
}

func NewMinerStatsRepository(db *gorm.DB) repository.MinerStatsRepository {
	return &minerStatsRepository{db: db}
}

func (r *minerStatsRepository) Save(ctx context.Context, stats *model.MinerStats) error {
	return r.db.WithContext(ctx).Create(stats).Error
}

func (r *minerStatsRepository) FindLatestByWorkerID(ctx context.Context, workerID string) (*model.MinerStats, error) {
	var stats model.MinerStats
	// Use Limit(1).Find to optimize query and avoid GORM's default PK ordering which might cause redundant sorting
	err := r.db.WithContext(ctx).Where("worker_id = ?", workerID).Order("id desc").Limit(1).Find(&stats).Error
	if err != nil {
		return nil, err
	}
	if stats.ID == 0 {
		return nil, nil
	}
	return &stats, nil
}
