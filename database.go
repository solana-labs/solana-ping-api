package main

import "time"

//PingResult schema for pingresult table
type PingResult struct {
	TimeStamp           int64  `gorm:"primaryKey;autoIncrement:false"`
	Cluster             string `gorm:"primaryKey"`
	Hostname            string
	Submitted           int
	Confirmed           int
	Loss                float64
	ConfirmationMessage string
	TakeTime            int
	Error               string    `gorm:"NOT NULL"`
	CreatedAt           time.Time `gorm:"type:timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP" json:"created_at,omitempty"`
	UpdatedAt           time.Time `gorm:"type:timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP" json:"updated_at,omitempty"`
}

func addRecord(data PingResult) error {
	dbMtx.Lock()
	result := database.Create(&data)
	dbMtx.Unlock()
	log.Info(data.Cluster, " add a record")
	return result.Error
}

func getLastN(c Cluster, n int) []PingResult {
	ret := []PingResult{}
	dbMtx.Lock()
	database.Order("time_stamp desc").Where("cluster=?", c).Limit(n).Find(&ret)
	dbMtx.Unlock()
	return ret
}

func getAfter(c Cluster, t int64) []PingResult {
	ret := []PingResult{}
	r := PingResult{}
	dbMtx.Lock()
	database.Order("time_stamp desc").First(&r)
	database.Order("time_stamp desc").Where("cluster=? AND time_stamp > ?", c, t).Find(&ret)
	log.Info("Latest in DB:", r.TimeStamp, " after:", t, " found:", len(ret))
	dbMtx.Unlock()
	return ret

}
