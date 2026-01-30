package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/beatyman/scan-miners/config"
	"github.com/beatyman/scan-miners/internal/domain/model"
	"github.com/beatyman/scan-miners/internal/domain/repository"
	"github.com/beatyman/scan-miners/pkg/logger"
	"github.com/beatyman/scan-miners/pkg/utils"
	"go.uber.org/zap"
)

type ScanWorkersUseCase struct {
	cfg        *config.Config
	workerRepo repository.WorkerRepository
	client     *http.Client
}

func NewScanWorkersUseCase(cfg *config.Config, repo repository.WorkerRepository) *ScanWorkersUseCase {
	return &ScanWorkersUseCase{
		cfg:        cfg,
		workerRepo: repo,
		client: &http.Client{
			Timeout: cfg.App.RequestTimeout,
		},
	}
}

func (uc *ScanWorkersUseCase) Execute(ctx context.Context) error {
	logger.Log.Info("Starting to scan workers from Antpool")

	// Prepare request
	// url := "https://www.antpool.com/auth/v3/observer/api/worker/list?search=&workerStatus=0&accessKey=tHRWhY0DJFTLgfPhE9tC&coinType=BTC&observerUserId=sam001sz&pageNum=1&pageSize=2000" // Increased pageSize to fetch more, assuming pagination logic if needed but user provided specific curl
	// Note: User provided pageSize=10 in curl, but we likely want all. Let's stick to the URL structure but maybe increase page size or loop pages.
	// For this task, I will fetch page 1 with a large page size to try to get many, or implement simple pagination if "totalPage" is returned.
	// The response shows "totalPage": 194, "totalRecord": 1932. So pageSize=10 is small.
	// I will implement loop to fetch all pages.

	page := 1
	pageSize := 100 // Reasonable batch size

	for {
		logger.Log.Info("Fetching workers page", zap.Int("page", page))

		reqURL := fmt.Sprintf("https://www.antpool.com/auth/v3/observer/api/worker/list?search=&workerStatus=0&accessKey=tHRWhY0DJFTLgfPhE9tC&coinType=BTC&observerUserId=sam001sz&pageNum=%d&pageSize=%d", page, pageSize)

		req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
		if err != nil {
			return err
		}

		// Set headers from requirement
		req.Header.Set("accept", "application/json, text/plain, */*")
		req.Header.Set("accept-language", "en,zh-CN;q=0.9,zh;q=0.8")
		req.Header.Set("cache-control", "no-cache")
		req.Header.Set("cookie", uc.cfg.App.AntpoolCookie)
		req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/144.0.0.0 Safari/537.36")
		// Add other headers if strictly necessary, but usually UA and Cookie are key.

		resp, err := uc.client.Do(req)
		if err != nil {
			logger.Log.Error("Failed to fetch workers", zap.Error(err))
			return err
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		// Parse Response
		var result struct {
			Code string `json:"code"`
			Msg  string `json:"msg"`
			Data struct {
				Items       []model.AntpoolWorkerResponseItem `json:"items"`
				PageNum     int                               `json:"pageNum"`
				TotalPage   int                               `json:"totalPage"`
				TotalRecord int                               `json:"totalRecord"`
			} `json:"data"`
		}

		if err := json.Unmarshal(body, &result); err != nil {
			logger.Log.Error("Failed to parse response", zap.Error(err))
			return err
		}

		if result.Code != "000000" {
			logger.Log.Error("API returned error", zap.String("code", result.Code), zap.String("msg", result.Msg))
			return fmt.Errorf("api error: %s", result.Msg)
		}

		// Convert and Save
		var workers []*model.Worker
		for _, item := range result.Data.Items {
			val10m, unit10m := utils.ParseHashrate(item.HsLast10Min)
			val1h, unit1h := utils.ParseHashrate(item.HsLast1H)
			val1d, unit1d := utils.ParseHashrate(item.HsLast1D)

			ip := utils.GenerateIP(item.WorkerID)

			worker := &model.Worker{
				WorkerID:          item.WorkerID,
				IP:                ip,
				UserWorkerID:      item.UserWorkerID,
				WorkerStatus:      item.WorkerStatus,
				HsLast10Min:       val10m,
				HsLast10MinUnit:   unit10m,
				HsLast1H:          val1h,
				HsLast1HUnit:      unit1h,
				HsLast1D:          val1d,
				HsLast1DUnit:      unit1d,
				RejectRatio:       item.RejectRatio,
				OnlineTimeLast24h: item.OnlineTimeLast24h,
				CreatedAt:         time.Unix(item.CreateTime/1000, 0), // Assuming ms timestamp
			}
			workers = append(workers, worker)
		}

		if len(workers) > 0 {
			if err := uc.workerRepo.SaveBatch(ctx, workers); err != nil {
				logger.Log.Error("Failed to save workers batch", zap.Error(err))
				return err
			}
			logger.Log.Info("Saved workers batch", zap.Int("count", len(workers)))
		}

		if page >= result.Data.TotalPage {
			break
		}
		page++
		time.Sleep(500 * time.Millisecond) // Polite delay
	}

	logger.Log.Info("Finished scanning workers")
	return nil
}
