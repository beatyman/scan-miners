package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/beatyman/scan-miners/config"
	"github.com/beatyman/scan-miners/internal/domain/model"
	"github.com/beatyman/scan-miners/internal/repository/mysql"
	"github.com/beatyman/scan-miners/internal/usecase"
	"github.com/beatyman/scan-miners/pkg/database"
	"github.com/beatyman/scan-miners/pkg/logger"
	"go.uber.org/zap"
)

func main() {
	// 1. Initialize Logger
	logger.Init()
	defer logger.Log.Sync()

	// 2. Parse Subcommands
	fetchWorkersCmd := flag.NewFlagSet("fetch-workers", flag.ExitOnError)
	scanMinersCmd := flag.NewFlagSet("scan-miners", flag.ExitOnError)
	exportAnalysisCmd := flag.NewFlagSet("export-analysis", flag.ExitOnError)
	exportUnderperformingCmd := flag.NewFlagSet("export-underperforming", flag.ExitOnError)

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	// 3. Common Setup (Config, DB, Repos)
	// We do this before switching commands because both need DB
	logger.Log.Info("Initializing Application...")
	cfg := config.Load()

	db, err := database.NewMySQLConnection(cfg)
	if err != nil {
		logger.Log.Fatal("Database connection failed", zap.Error(err))
	}

	logger.Log.Info("Running database migrations...")
	if err := db.AutoMigrate(&model.Worker{}, &model.MinerStats{}, &model.MinerChain{}); err != nil {
		logger.Log.Fatal("Migration failed", zap.Error(err))
	}

	workerRepo := mysql.NewWorkerRepository(db)
	minerStatsRepo := mysql.NewMinerStatsRepository(db)

	scanWorkersUC := usecase.NewScanWorkersUseCase(cfg, workerRepo)
	scanMinersUC := usecase.NewScanMinersUseCase(cfg, workerRepo, minerStatsRepo)
	exportAnalysisUC := usecase.NewExportHashrateAnalysisUseCase(workerRepo, minerStatsRepo)
	exportUnderperformingUC := usecase.NewExportUnderperformingMinersUseCase(workerRepo, minerStatsRepo)
	ctx := context.Background()

	// 4. Execute Logic based on Subcommand
	switch os.Args[1] {
	case "fetch-workers":
		fetchWorkersCmd.Parse(os.Args[2:])
		logger.Log.Info(">>> Executing: Fetch Workers from Antpool <<<")
		if err := scanWorkersUC.Execute(ctx); err != nil {
			logger.Log.Fatal("Fetch workers failed", zap.Error(err))
		}
	case "scan-miners":
		scanMinersCmd.Parse(os.Args[2:])
		logger.Log.Info(">>> Executing: Scan Miner Stats <<<")
		if err := scanMinersUC.Execute(ctx); err != nil {
			logger.Log.Fatal("Scan miners failed", zap.Error(err))
		}
	case "export-analysis":
		exportAnalysisCmd.Parse(os.Args[2:])
		logger.Log.Info(">>> Executing: Export Hashrate Analysis <<<")
		if err := exportAnalysisUC.Execute(ctx); err != nil {
			logger.Log.Fatal("Export analysis failed", zap.Error(err))
		}
	case "export-underperforming":
		exportUnderperformingCmd.Parse(os.Args[2:])
		logger.Log.Info(">>> Executing: Export Underperforming Miners <<<")
		if err := exportUnderperformingUC.Execute(ctx); err != nil {
			logger.Log.Fatal("Export underperforming failed", zap.Error(err))
		}
	default:
		printUsage()
		os.Exit(1)
	}

	logger.Log.Info("Task completed successfully.")
}

func printUsage() {
	fmt.Println("Usage: sacn-miners <subcommand> [options]")
	fmt.Println("\nSubcommands:")
	fmt.Println("  fetch-workers    Fetch worker list from Antpool and save to DB")
	fmt.Println("  scan-miners      Scan miner stats using IPs from DB")
	fmt.Println("  export-analysis  Export hashrate analysis to CSV")
	fmt.Println("  export-underperforming  Export miners with hashrate below rated value")
	fmt.Println("")
}
