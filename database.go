package main

import (
	"time"

	"github.com/lib/pq"
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
	dbMtx.Lock()
	result := database.Create(&data)
	dbMtx.Unlock()
	return result.Error
}

func getLastN(c Cluster, pType PingType, n int) []PingResult {
	ret := []PingResult{}
	dbMtx.Lock()
	database.Order("time_stamp desc").Where("cluster=? AND ping_type=?", c, string(pType)).Limit(n).Find(&ret)
	dbMtx.Unlock()
	return ret
}
func getAfter(c Cluster, pType PingType, t int64) []PingResult {
	ret := []PingResult{}
	now := time.Now().UTC().Unix()
	dbMtx.Lock()
	database.Where("cluster=? AND ping_type=? AND time_stamp > ? AND time_stamp < ?", c, string(pType), t, now).Find(&ret)
	dbMtx.Unlock()
	return ret
}

func deleteTimeBefore(t int64) {
	database.Where("time_stamp < ?", t).Delete(&[]PingResult{})
}
