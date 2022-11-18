package main

import (
	"context"
	"math"
	"strings"
	"time"

	"github.com/portto/solana-go-sdk/client"
	"github.com/portto/solana-go-sdk/types"
)

// TakeTime a struct to record a serious of start and end time. Start and End is used for current record and Times is for store take-time
type TakeTime struct {
	Times []int64
	Start int64
	End   int64
}

// Ping similar to solana-bench-tps. It send a transaction to the cluster
func Ping(c *client.Client, pType PingType, acct types.Account, config ClusterConfig) (PingResult, PingResultError) {
	resultErrs := []string{}
	timer := TakeTime{}
	result := PingResult{
		Cluster:  string(config.Cluster),
		Hostname: config.HostName,
		PingType: string(pType),
	}
	confirmedCount := 0
	for i := 0; i < config.BatchCount; i++ {
		if i > 0 {
			time.Sleep(time.Duration(config.BatchInverval))
		}
		timer.TimerStart()
		var hash string
		if 0 == config.ComputeUnitPrice {
			txhash, pingErr := Transfer(c, acct, acct, config.Receiver, time.Duration(config.TxTimeout)*time.Second)
			hash = txhash // avoid shadow
			if !pingErr.NoError() {
				timer.TimerStop()
				if !pingErr.IsInErrorList(PingTakeTimeErrExpectionList) {
					timer.Add()
				}
				resultErrs = append(resultErrs, string(pingErr))
				continue
			}
		} else {
			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(config.TxTimeout)*time.Second)
			defer cancel()
			txhash, pingErr := SendPingTx(SendPingTxParam{
				Client:              c,
				Ctx:                 ctx,
				FeePayer:            acct,
				RequestComputeUnits: config.RequestUnits,
				ComputeUnitPrice:    config.ComputeUnitPrice,
			})
			hash = txhash // avoid shadow
			if !pingErr.NoError() {
				timer.TimerStop()
				if !pingErr.IsInErrorList(PingTakeTimeErrExpectionList) {
					timer.Add()
				}
				resultErrs = append(resultErrs, string(pingErr))
				continue
			}
		}
		pingErr := waitConfirmation(c, hash,
			time.Duration(config.WaitConfirmationTimeout)*time.Second,
			time.Duration(config.TxTimeout)*time.Second,
			time.Duration(config.StatusCheckInterval)*time.Second)
		timer.TimerStop()
		if !pingErr.NoError() {
			resultErrs = append(resultErrs, string(pingErr))
			if !pingErr.IsInErrorList(PingTakeTimeErrExpectionList) {
				timer.Add()
			}
			continue
		}
		timer.Add()
		confirmedCount++
		// log.Println(hash, "is confirmed/finalized")

	}
	result.TimeStamp = time.Now().UTC().Unix()
	result.Submitted = config.BatchCount
	result.Confirmed = confirmedCount
	result.Loss = (float64(result.Submitted-result.Confirmed) / float64(result.Submitted)) * 100
	max, mean, min, stdDev, total := timer.Statistic()
	result.Max = max
	result.Mean = int64(mean)
	result.Min = min
	result.Stddev = int64(stdDev)
	result.TakeTime = total
	result.ComputeUnitPrice = config.ComputeUnitPrice
	result.RequestComputeUnits = config.RequestUnits
	result.Error = resultErrs
	stringErrors := []string(result.Error)
	if 0 == len(stringErrors) {
		return result, EmptyPingResultError
	}
	return result, PingResultError(strings.Join(stringErrors[:], ","))
}

// TimerStart Record start time in ms format
func (t *TakeTime) TimerStart() {
	t.Start = time.Now().UTC().UnixMilli()
}

// TimerStop Record stop time in ms format
func (t *TakeTime) TimerStop() {
	t.End = time.Now().UTC().UnixMilli()
}

// Add save the end - stop in ms
func (t *TakeTime) Add() {
	t.Times = append(t.Times, (t.End - t.Start))
}

// AddTime add a take time directly into Times
func (t *TakeTime) AddTime(ts int64) {
	t.Times = append(t.Times, ts)
}

// TotalTime sum of data in Times
func (t *TakeTime) TotalTime() int64 {
	sum := int64(0)
	for _, ts := range t.Times {
		sum += ts
	}
	return sum
}

// Statistic analyze data in TakeTime to return max/mean/min/stddev/sum
func (t *TakeTime) Statistic() (max int64, mean float64, min int64, stddev float64, sum int64) {
	count := 0
	for _, ts := range t.Times {
		if ts <= 0 { // do not use 0 data because it is the bad data
			continue
		}
		if max == 0 {
			max = ts
		}
		if min == 0 {
			min = ts
		}
		if ts >= max {
			max = ts
		}
		if ts <= min {
			min = ts
		}
		sum += ts
		count++

	}
	if count > 0 {
		mean = float64(sum) / float64(count)
		for _, ts := range t.Times {
			if ts > 0 { // if ts = 0 , ping fail.
				stddev += math.Pow(float64(ts)-mean, 2)
			}
		}
		stddev = math.Sqrt(stddev / float64(count))
	}
	return
}
