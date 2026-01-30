package usecase

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/beatyman/scan-miners/internal/domain/repository"
	"github.com/beatyman/scan-miners/pkg/logger"
	"go.uber.org/zap"
)

type ExportUnderperformingMinersUseCase struct {
	workerRepo     repository.WorkerRepository
	minerStatsRepo repository.MinerStatsRepository
}

func NewExportUnderperformingMinersUseCase(workerRepo repository.WorkerRepository, minerStatsRepo repository.MinerStatsRepository) *ExportUnderperformingMinersUseCase {
	return &ExportUnderperformingMinersUseCase{
		workerRepo:     workerRepo,
		minerStatsRepo: minerStatsRepo,
	}
}

// Rated hashrates mapping (TH/s)
var ratedHashrates = map[string]float64{
	"Antminer U3S19XP+H (HashMaster)":   293,
	"Antminer U3S19XP+H Ex":             293,
	"Antminer U3S19EXPH (HashMaster)":   251,
	"Antminer U3S19XP+H":                279,
	"Antminer U3S19EXPH":                251,
	"Antminer S19 XP+ Hyd (HashMaster)": 293,
	"Antminer S19 XP+ Hyd.":             293,
	"Antminer S19e XP Hyd Ex":           279,
}

func (uc *ExportUnderperformingMinersUseCase) Execute(ctx context.Context) error {
	logger.Log.Info("Starting underperforming miners export")

	// 1. Fetch all workers
	workers, err := uc.workerRepo.FindAll(ctx)
	if err != nil {
		return err
	}

	logger.Log.Info("Found workers to analyze", zap.Int("count", len(workers)))

	// 2. Prepare CSV file
	filename := fmt.Sprintf("underperforming_miners_%s.csv", time.Now().Format("20060102_150405"))
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Add BOM for Excel compatibility
	file.Write([]byte{0xEF, 0xBB, 0xBF})

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// 3. Write Header
	// IP, rate_avg, rate_ideal, 额定算力, 差值
	header := []string{
		"IP",
		"Miner Type",
		"Rate Avg (TH/s)",
		"Rate Ideal (TH/s)",
		"Rated Hashrate (TH/s)",
		"Difference (TH/s)",
	}
	if err := writer.Write(header); err != nil {
		return err
	}

	count := 0
	// 4. Process each worker
	for _, worker := range workers {
		stats, err := uc.minerStatsRepo.FindLatestByWorkerID(ctx, worker.WorkerID)
		if err != nil || stats == nil {
			continue
		}

		// Normalize miner type (trim spaces)
		minerType := strings.TrimSpace(stats.MinerType)

		// Find rated hashrate
		rated, ok := ratedHashrates[minerType]
		if !ok {
			// Try fuzzy match or skip?
			// For now, let's try to match exactly as provided.
			// If not found, we can't determine if underperforming based on rated.
			// Maybe log warning?
			// logger.Log.Warn("Unknown miner type for rated hashrate", zap.String("type", minerType))
			continue
		}

		rateAvg := convertToTHs(stats.RateAvg, stats.RateUnit)
		rateIdeal := convertToTHs(stats.RateIdeal, stats.RateUnit)

		// Check if rate_avg < rated
		if rateAvg < rated {
			diff := rateAvg - rated
			record := []string{
				worker.IP,
				minerType,
				fmt.Sprintf("%.2f", rateAvg),
				fmt.Sprintf("%.2f", rateIdeal),
				fmt.Sprintf("%.2f", rated),
				fmt.Sprintf("%.2f", diff),
			}
			if err := writer.Write(record); err != nil {
				return err
			}
			count++
		}
	}

	absPath, _ := filepath.Abs(filename)
	logger.Log.Info("Export completed successfully", zap.String("file", absPath), zap.Int("underperforming_count", count))
	return nil
}
