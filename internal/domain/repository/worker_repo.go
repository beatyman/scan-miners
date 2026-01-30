package repository

import (
	"context"
	"github.com/beatyman/scan-miners/internal/domain/model"
)

type WorkerRepository interface {
	Save(ctx context.Context, worker *model.Worker) error
	SaveBatch(ctx context.Context, workers []*model.Worker) error
	FindAll(ctx context.Context) ([]*model.Worker, error)
	FindByWorkerID(ctx context.Context, workerID string) (*model.Worker, error)
}
