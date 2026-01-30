package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/beatyman/scan-miners/config"
	"github.com/beatyman/scan-miners/internal/domain/model"
	"github.com/beatyman/scan-miners/internal/domain/repository"
	"github.com/beatyman/scan-miners/pkg/logger"
	"github.com/icholy/digest"
	"go.uber.org/zap"
)

type ScanMinersUseCase struct {
	cfg            *config.Config
	workerRepo     repository.WorkerRepository
	minerStatsRepo repository.MinerStatsRepository
	client         *http.Client
}

func NewScanMinersUseCase(cfg *config.Config, workerRepo repository.WorkerRepository, minerStatsRepo repository.MinerStatsRepository) *ScanMinersUseCase {
	// Setup digest authentication client
	t := &digest.Transport{
		Username: cfg.App.MinerUser,
		Password: cfg.App.MinerPassword,
	}

	return &ScanMinersUseCase{
		cfg:            cfg,
		workerRepo:     workerRepo,
		minerStatsRepo: minerStatsRepo,
		client: &http.Client{
			Transport: t,
			Timeout:   5 * time.Second, // Shorter timeout for local network scanning
		},
	}
}

func (uc *ScanMinersUseCase) Execute(ctx context.Context) error {
	logger.Log.Info("Starting to scan miner stats")

	workers, err := uc.workerRepo.FindAll(ctx)
	if err != nil {
		return err
	}

	logger.Log.Info("Found workers to scan", zap.Int("count", len(workers)))

	// Worker pool for scanning
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 50) // Limit concurrency to 50

	for _, worker := range workers {
		if worker.IP == "" {
			continue
		}

		wg.Add(1)
		semaphore <- struct{}{}

		go func(w *model.Worker) {
			defer wg.Done()
			defer func() { <-semaphore }()

			if err := uc.scanSingleMiner(ctx, w); err != nil {
				// Log error but continue
				logger.Log.Debug("Failed to scan miner", zap.String("ip", w.IP), zap.Error(err))
			}
		}(worker)
	}

	wg.Wait()
	logger.Log.Info("Finished scanning miner stats")
	return nil
}

func (uc *ScanMinersUseCase) scanSingleMiner(ctx context.Context, worker *model.Worker) error {
	endpoints := []string{
		fmt.Sprintf("http://%s/cgi-bin/get_stats.cgi", worker.IP),
		fmt.Sprintf("http://%s/cgi-bin/stats.cgi", worker.IP),
	}

	var body []byte
	var err error

	for _, url := range endpoints {
		body, err = uc.fetchURL(ctx, url)
		if err == nil {
			break
		}
	}

	if err != nil {
		return err
	}

	// Parse JSON
	var resp model.MinerAPIResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return err
	}

	// Transform to Domain Model
	// Assuming only one STATS item as per usual Antminer API
	if len(resp.Stats) == 0 {
		return fmt.Errorf("no stats data found for %s", worker.IP)
	}

	statItem := resp.Stats[0]

	minerStats := &model.MinerStats{
		WorkerID:     worker.WorkerID,
		IP:           worker.IP,
		MinerType:    resp.Info.Type,
		MinerVersion: resp.Info.MinerVersion,
		CompileTime:  resp.Info.CompileTime,
		Elapsed:      statItem.Elapsed,
		Rate5s:       statItem.Rate5s,
		Rate30m:      statItem.Rate30m,
		RateAvg:      statItem.RateAvg,
		RateIdeal:    statItem.RateIdeal,
		RateUnit:     statItem.RateUnit,
		FanNum:       statItem.FanNum,
		HwpTotal:     statItem.HwpTotal,
	}

	// Map Chains
	for _, chainItem := range statItem.Chain {
		minerStats.Chains = append(minerStats.Chains, model.MinerChain{
			ChainIndex: chainItem.Index,
			FreqAvg:    chainItem.FreqAvg,
			RateIdeal:  chainItem.RateIdeal,
			RateReal:   chainItem.RateReal,
			AsicNum:    chainItem.AsicNum,
			Hw:         chainItem.Hw,
			Hwp:        chainItem.Hwp,
		})
	}

	// Save
	if err := uc.minerStatsRepo.Save(ctx, minerStats); err != nil {
		return err
	}

	logger.Log.Info("Successfully scanned miner", zap.String("ip", worker.IP))
	return nil
}

func (uc *ScanMinersUseCase) fetchURL(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := uc.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status: %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}
