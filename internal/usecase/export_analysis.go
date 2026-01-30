package usecase

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/beatyman/scan-miners/internal/domain/repository"
	"github.com/beatyman/scan-miners/pkg/logger"
	"go.uber.org/zap"
)

type ExportHashrateAnalysisUseCase struct {
	workerRepo     repository.WorkerRepository
	minerStatsRepo repository.MinerStatsRepository
}

func NewExportHashrateAnalysisUseCase(workerRepo repository.WorkerRepository, minerStatsRepo repository.MinerStatsRepository) *ExportHashrateAnalysisUseCase {
	return &ExportHashrateAnalysisUseCase{
		workerRepo:     workerRepo,
		minerStatsRepo: minerStatsRepo,
	}
}

func (uc *ExportHashrateAnalysisUseCase) Execute(ctx context.Context) error {
	logger.Log.Info("Starting hashrate analysis export")

	// 1. Fetch all workers
	workers, err := uc.workerRepo.FindAll(ctx)
	if err != nil {
		return err
	}

	logger.Log.Info("Found workers to analyze", zap.Int("count", len(workers)))

	// 2. Prepare CSV file
	filename := fmt.Sprintf("hashrate_analysis_%s.csv", time.Now().Format("20060102_150405"))
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Add BOM for Excel compatibility (optional but good for UTF-8)
	file.Write([]byte{0xEF, 0xBB, 0xBF})

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// 3. Write Header
	// iP , miner_type ,rate_avg ,rate_ideal, hs_last_1d ,reject_ratio ,worker_status,online_time_last24h
	header := []string{
		"IP",
		"Miner Type",
		"Rate Avg (TH/s)",
		"Rate Ideal (TH/s)",
		"Hs Last 1D (TH/s)",
		"Reject Ratio",
		"Worker Status",
		"Online Time Last 24h",
	}
	if err := writer.Write(header); err != nil {
		return err
	}

	// 4. Process each worker
	for _, worker := range workers {
		// Fetch latest miner stats
		var minerType string
		var rateAvg, rateIdeal float64
		
		// If worker has IP, try to fetch local stats from DB
		// Since we stored them by WorkerID, we use that
		stats, err := uc.minerStatsRepo.FindLatestByWorkerID(ctx, worker.WorkerID)
		if err == nil && stats != nil {
			minerType = stats.MinerType
			rateAvg = convertToTHs(stats.RateAvg, stats.RateUnit)
			rateIdeal = convertToTHs(stats.RateIdeal, stats.RateUnit)
		} else {
			// If no stats found, just leave empty or zero
			// logger.Log.Debug("No miner stats found for worker", zap.String("workerID", worker.WorkerID))
		}

		// Convert Worker stats
		hsLast1d := convertToTHs(worker.HsLast1D, worker.HsLast1DUnit)

		record := []string{
			worker.IP,
			minerType,
			fmt.Sprintf("%.2f", rateAvg),
			fmt.Sprintf("%.2f", rateIdeal),
			fmt.Sprintf("%.2f", hsLast1d),
			worker.RejectRatio,
			strconv.Itoa(worker.WorkerStatus),
			fmt.Sprintf("%.2f", worker.OnlineTimeLast24h),
		}

		if err := writer.Write(record); err != nil {
			return err
		}
	}

	absPath, _ := filepath.Abs(filename)
	logger.Log.Info("Export completed successfully", zap.String("file", absPath))
	return nil
}

func convertToTHs(value float64, unit string) float64 {
	if value == 0 {
		return 0
	}
	
	// Normalize unit
	unit = strings.TrimSpace(unit)
	unit = strings.ToUpper(unit)

	switch unit {
	case "TH/S", "TH":
		return value
	case "GH/S", "GH":
		return value / 1000
	case "MH/S", "MH":
		return value / 1000000
	case "PH/S", "PH":
		return value * 1000
	default:
		// Default assumption or no conversion if unit unknown
		// But usually Antminer returns GH/s or TH/s
		// If unit is empty, we assume it's raw value, check magnitude? 
		// For now, let's assume if unit is missing it might be GH/s (common in older firmware) or TH/s.
		// Let's assume TH/s if unsure or return raw.
		// Actually, let's look at sample data: "hsLast10Min": "306.23 TH/s".
		// Stats API: "rate_unit": "GH/s".
		return value
	}
}
