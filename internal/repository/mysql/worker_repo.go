package mysql

import (
	"context"
	"github.com/beatyman/scan-miners/internal/domain/model"
	"github.com/beatyman/scan-miners/internal/domain/repository"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type workerRepository struct {
	db *gorm.DB
}

func NewWorkerRepository(db *gorm.DB) repository.WorkerRepository {
	return &workerRepository{db: db}
}

func (r *workerRepository) Save(ctx context.Context, worker *model.Worker) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "worker_id"}},
		UpdateAll: true,
	}).Create(worker).Error
}

func (r *workerRepository) SaveBatch(ctx context.Context, workers []*model.Worker) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "worker_id"}},
		UpdateAll: true,
	}).CreateInBatches(workers, 100).Error
}

func (r *workerRepository) FindAll(ctx context.Context) ([]*model.Worker, error) {
	var workers []*model.Worker
	err := r.db.WithContext(ctx).Find(&workers).Error
	return workers, err
}

func (r *workerRepository) FindByWorkerID(ctx context.Context, workerID string) (*model.Worker, error) {
	var worker model.Worker
	err := r.db.WithContext(ctx).Where("worker_id = ?", workerID).First(&worker).Error
	if err != nil {
		return nil, err
	}
	return &worker, nil
}
