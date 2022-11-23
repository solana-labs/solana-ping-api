package main

import (
	"time"

	"github.com/lib/pq"
)

// ComputeUnitPriceType tell program fetch what kind of compute data to fetch
type ComputeUnitPriceType string

// Cluster enum
const (
	AllData                   ComputeUnitPriceType = "all"
	NoComputeUnitPrice                             = "zero"
	HasComputeUnitPrice                            = "hasprice"
	ComputeUnitPriceThreshold                      = "threshold"
)

// PingResult is a struct to store ping result and database structure
type PingResult struct {
	TimeStamp           int64 `gorm:"autoIncrement:false"`
	Cluster             string
	Hostname            string
	PingType            string `gorm:"NOT NULL"`
	Submitted           int    `gorm:"NOT NULL"`
	Confirmed           int    `gorm:"NOT NULL"`
	Loss                float64
	Max                 int64
	Mean                int64
	Min                 int64
	Stddev              int64
	TakeTime            int64
	RequestComputeUnits uint32
	ComputeUnitPrice    uint64
	Error               pq.StringArray `gorm:"type:text[];"NOT NULL"`
	CreatedAt           time.Time      `gorm:"type:timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP" json:"created_at,omitempty"`
	UpdatedAt           time.Time      `gorm:"type:timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP" json:"updated_at,omitempty"`
}

func addRecord(data PingResult) error {
	result := database.Create(&data)
	return result.Error
}

func getLastN(c Cluster, pType PingType, n int, priceType ComputeUnitPriceType, threshold uint64) []PingResult {
	ret := []PingResult{}
	switch priceType {
	case NoComputeUnitPrice:
		database.Order("time_stamp desc").Where("cluster=? AND ping_type=? AND compute_unit_price = ?", c, string(pType), 0).Limit(n).Find(&ret)
	case HasComputeUnitPrice:
		database.Order("time_stamp desc").Where("cluster=? AND ping_type=? AND compute_unit_price > ?", c, string(pType), 0).Limit(n).Find(&ret)
	case ComputeUnitPriceThreshold:
		database.Order("time_stamp desc").Where("cluster=? AND ping_type=? AND compute_unit_price > ?", c, string(pType), threshold).Limit(n).Find(&ret)
	case AllData:
		fallthrough
	default:
		database.Order("time_stamp desc").Where("cluster=? AND ping_type=?", c, string(pType)).Limit(n).Find(&ret)
	}
	return ret
}
func getAfter(c Cluster, pType PingType, t int64, priceType ComputeUnitPriceType, threshold uint64) []PingResult {
	ret := []PingResult{}
	now := time.Now().UTC().Unix()
	switch priceType {
	case NoComputeUnitPrice:
		database.Where("cluster=? AND ping_type=? AND time_stamp > ? AND time_stamp < ? AND compute_unit_price = ?", c, string(pType), t, now, 0).Find(&ret)
	case HasComputeUnitPrice:
		database.Where("cluster=? AND ping_type=? AND time_stamp > ? AND time_stamp < ? AND compute_unit_price > ?", c, string(pType), t, now, 0).Find(&ret)
	case ComputeUnitPriceThreshold:
		database.Where("cluster=? AND ping_type=? AND time_stamp > ? AND time_stamp < ? AND compute_unit_price > ?", c, string(pType), t, now, threshold).Find(&ret)
	case AllData:
		fallthrough
	default:
		database.Where("cluster=? AND ping_type=? AND time_stamp > ? AND time_stamp < ?", c, string(pType), t, now).Find(&ret)
	}

	return ret
}

func deleteTimeBefore(t int64) {
	database.Where("time_stamp < ?", t).Delete(&[]PingResult{})
}
