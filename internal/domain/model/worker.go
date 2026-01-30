package model

import (
	"time"
)

type Worker struct {
	ID                uint      `gorm:"primaryKey" json:"-"`
	WorkerID          string    `gorm:"uniqueIndex;type:varchar(64)" json:"workerId"`
	IP                string    `gorm:"type:varchar(64)" json:"ip"`
	UserWorkerID      string    `gorm:"type:varchar(128)" json:"userWorkerId"`
	WorkerStatus      int       `json:"workerStatus"`
	
	HsLast10Min       float64   `json:"hsLast10Min"`
	HsLast10MinUnit   string    `gorm:"type:varchar(16)" json:"hsLast10MinUnit"`
	
	HsLast1H          float64   `json:"hsLast1H"`
	HsLast1HUnit      string    `gorm:"type:varchar(16)" json:"hsLast1HUnit"`
	
	HsLast1D          float64   `json:"hsLast1D"`
	HsLast1DUnit      string    `gorm:"type:varchar(16)" json:"hsLast1DUnit"`
	
	RejectRatio       string    `gorm:"type:varchar(16)" json:"rejectRatio"`
	OnlineTimeLast24h float64   `json:"onlineTimeLast24h"`
	
	CreatedAt         time.Time `json:"createTime"` // Antpool returns timestamp, we might need custom unmarshaler or handle logic
	UpdatedAt         time.Time
}

// AntpoolWorkerResponseItem is used for JSON unmarshalling from Antpool API
type AntpoolWorkerResponseItem struct {
	ID                int64   `json:"id"`
	WorkerID          string  `json:"workerId"`
	UserWorkerID      string  `json:"userWorkerId"`
	WorkerStatus      int     `json:"workerStatus"`
	HsLast10Min       string  `json:"hsLast10Min"`
	HsLast1H          string  `json:"hsLast1H"` // Note: API has hsLast1Hour and hsLast1H
	HsLast1D          string  `json:"hsLast1D"`
	RejectRatio       string  `json:"rejectRatio"`
	OnlineTimeLast24h float64 `json:"onlineTimeLast24h"`
	CreateTime        int64   `json:"createTime"`
}
