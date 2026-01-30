package model

import (
	"time"
)

type MinerStats struct {
	ID           uint   `gorm:"primaryKey;index:idx_worker_latest,priority:2,sort:desc"`
	WorkerID     string `gorm:"type:varchar(64);index:idx_worker_latest,priority:1"` // Logically linked to Worker
	IP           string `gorm:"type:varchar(64)"`
	MinerType    string `gorm:"type:varchar(64)"`
	MinerVersion string `gorm:"type:varchar(64)"`
	CompileTime  string `gorm:"type:varchar(64)"`

	Elapsed   int64
	Rate5s    float64
	Rate30m   float64
	RateAvg   float64
	RateIdeal float64
	RateUnit  string `gorm:"type:varchar(16)"`
	FanNum    int
	HwpTotal  float64

	Chains []MinerChain `gorm:"foreignKey:MinerStatsID"`

	CreatedAt time.Time
}

type MinerChain struct {
	ID           uint `gorm:"primaryKey"`
	MinerStatsID uint `gorm:"index"`
	ChainIndex   int
	FreqAvg      int
	RateIdeal    float64
	RateReal     float64
	AsicNum      int
	Hw           int
	Hwp          float64
	// We simplify temp arrays to store avg or just raw string if needed,
	// for now let's skip complex arrays or just store basic info as per requirement "解析到二层就就可以"

	CreatedAt time.Time
}

// MinerAPIResponse structures for JSON unmarshalling
type MinerAPIResponse struct {
	Status map[string]interface{} `json:"STATUS"`
	Info   MinerInfo              `json:"INFO"`
	Stats  []MinerStatItem        `json:"STATS"`
}

type MinerInfo struct {
	MinerVersion string `json:"miner_version"`
	CompileTime  string `json:"CompileTime"`
	Type         string `json:"type"`
}

type MinerStatItem struct {
	Elapsed   int64            `json:"elapsed"`
	Rate5s    float64          `json:"rate_5s"`
	Rate30m   float64          `json:"rate_30m"`
	RateAvg   float64          `json:"rate_avg"`
	RateIdeal float64          `json:"rate_ideal"`
	RateUnit  string           `json:"rate_unit"`
	ChainNum  int              `json:"chain_num"`
	FanNum    int              `json:"fan_num"`
	HwpTotal  float64          `json:"hwp_total"`
	Chain     []MinerChainItem `json:"chain"`
}

type MinerChainItem struct {
	Index     int     `json:"index"`
	FreqAvg   int     `json:"freq_avg"`
	RateIdeal float64 `json:"rate_ideal"`
	RateReal  float64 `json:"rate_real"`
	AsicNum   int     `json:"asic_num"`
	Hw        int     `json:"hw"`
	Hwp       float64 `json:"hwp"`
}
