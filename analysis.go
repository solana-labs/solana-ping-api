package main

import (
	"fmt"
	"log"
	"time"
)

// collect ping result and divided them into a group by each min
type Group1Min struct {
	Result    []PingResult
	TimeStamp int64
}

// statistic of ping result take-time
type TimeStatistic struct {
	Min    int64
	Mean   float64
	Max    int64
	Stddev float64
	Sum    int64
}

// statistic of ping result
type PingSatistic struct {
	Submitted   float64
	Confirmed   float64
	Loss        float64
	Count       int
	TimeMeasure TakeTime
	TimeStatistic
	Errors    []string
	TimeStamp int64
}

// statistic without a group
type GlobalStatistic struct {
	Submitted float64
	Confirmed float64
	Loss      float64
	Count     int
	TimeStatistic
}

// All statistic Data
type GroupsAllStatistic struct {
	PingStatisticList    []PingSatistic
	RawPingStaticList    []PingSatistic
	GlobalErrorStatistic map[string]int
	GlobalStatistic
}

func (g *GroupsAllStatistic) GetGroupsAllStatistic(raw bool) GlobalStatistic {
	groupStat := GlobalStatistic{}
	var sumOfSubmitted, sumOfConfirmed float64
	var sumOfCount int
	var sumTimeMeasure TakeTime
	if !raw {
		for _, pg := range g.PingStatisticList {
			sumOfSubmitted += pg.Submitted
			sumOfConfirmed += pg.Confirmed
			sumOfCount += pg.Count
			sumTimeMeasure.Times = append(sumTimeMeasure.Times, pg.TimeMeasure.Times...)
		}

	} else { // raw data (without exception)
		for _, pg := range g.RawPingStaticList {
			sumOfSubmitted += pg.Submitted
			sumOfConfirmed += pg.Confirmed
			sumOfCount += pg.Count
			sumTimeMeasure.Times = append(sumTimeMeasure.Times, pg.TimeMeasure.Times...)

		}
	}

	groupStat.Submitted = sumOfSubmitted
	groupStat.Confirmed = sumOfConfirmed
	groupStat.Loss = 1
	if groupStat.Submitted > 0 {
		groupStat.Loss = (groupStat.Submitted - groupStat.Confirmed) / groupStat.Submitted
	} else if len(g.GlobalErrorStatistic) == 0 { // No Data
		groupStat.Loss = 0
	}

	tMax, tMean, tMin, tStddev, tSum := sumTimeMeasure.Statistic()
	groupStat.TimeStatistic = TimeStatistic{
		Min:    tMin,
		Mean:   tMean,
		Max:    tMax,
		Stddev: tStddev,
		Sum:    tSum,
	}
	return groupStat
}

// grouping1Min: group []PingResult into 1 min group.
func grouping1Min(pr []PingResult, startTime int64, endTime int64) []Group1Min {
	window := int64(60)
	groups := []Group1Min{}
	for periodend := endTime; periodend > startTime; periodend = periodend - window {
		minGroup := Group1Min{}
		for _, pResult := range pr {
			if pResult.TimeStamp <= periodend && pResult.TimeStamp > periodend-window {
				minGroup.Result = append(minGroup.Result, pResult)
			}
		}
		minGroup.TimeStamp = periodend
		// debug: printPingResultGroup(minGroup.Result, periodend, periodend-window)
		groups = append(groups, minGroup)
	}
	return groups
}

func statisticCompute(cConf ClusterConfig, groups []Group1Min) *GroupsAllStatistic {
	stat := GroupsAllStatistic{}
	stat.PingStatisticList = []PingSatistic{}
	stat.RawPingStaticList = []PingSatistic{}
	stat.GlobalErrorStatistic = make(map[string]int)

	for _, group := range groups {
		filterGroupStat := PingSatistic{}
		rawGroupStat := PingSatistic{}
		for _, singlePing := range group.Result {
			errorException := false
			errorCount := len(singlePing.Error)
			if errorCount > 0 {
				for _, e := range singlePing.Error {
					stat.GlobalErrorStatistic[string(e)] = stat.GlobalErrorStatistic[string(e)] + 1
					if PingResultError(e).IsInErrorList(StatisticErrorExceptionList) {
						errorException = true
					} else {
						filterGroupStat.Errors = append(filterGroupStat.Errors, string(e))
					}
					rawGroupStat.Errors = append(rawGroupStat.Errors, string(e))
				}
			}
			// Raw Data Statistic
			rawGroupStat.TimeStamp = group.TimeStamp
			rawGroupStat.Submitted += float64(singlePing.Submitted)
			rawGroupStat.Confirmed += float64(singlePing.Confirmed)
			rawGroupStat.Count += 1
			rawGroupStat.TimeMeasure.AddTime(singlePing.TakeTime)
			// Data Statistic (Filtered by error filter)
			filterGroupStat.TimeStamp = group.TimeStamp // Need a ts to present the group
			if !errorException {
				filterGroupStat.Submitted += float64(singlePing.Submitted)
				filterGroupStat.Confirmed += float64(singlePing.Confirmed)
				filterGroupStat.Count += 1
				if errorCount <= 0 {
					filterGroupStat.TimeMeasure.AddTime(singlePing.TakeTime)
				} else if (errorCount > 0) && !errorException { // general error is considered as a timeout
					t := time.Duration(cConf.PingConfig.TxTimeout) * time.Second
					filterGroupStat.TimeMeasure.AddTime(t.Milliseconds())
				} // if StatisticErrorExceptionList , do not count as a satistic
			}
		}
		// raw data
		if rawGroupStat.Submitted == 0 { // no data
			rawGroupStat.Loss = 1
		} else {
			rawGroupStat.Loss = (rawGroupStat.Submitted - rawGroupStat.Confirmed) / rawGroupStat.Submitted
		}
		tMax, tMean, tMin, tStddev, tSum := rawGroupStat.TimeMeasure.Statistic()
		rawGroupStat.TimeStatistic = TimeStatistic{
			Min:    tMin,
			Mean:   tMean,
			Max:    tMax,
			Stddev: tStddev,
			Sum:    tSum,
		}
		stat.RawPingStaticList = append(stat.RawPingStaticList, rawGroupStat)

		// data with filter
		if filterGroupStat.Submitted == 0 { // no data
			filterGroupStat.Loss = 1
		} else {
			filterGroupStat.Loss = (filterGroupStat.Submitted - filterGroupStat.Confirmed) / filterGroupStat.Submitted
		}

		tMax, tMean, tMin, tStddev, tSum = filterGroupStat.TimeMeasure.Statistic()
		filterGroupStat.TimeStatistic = TimeStatistic{
			Min:    tMin,
			Mean:   tMean,
			Max:    tMax,
			Stddev: tStddev,
			Sum:    tSum,
		}
		stat.PingStatisticList = append(stat.PingStatisticList, filterGroupStat)
	}
	return &stat
}

func printPingResultGroup(pr []PingResult, from int64, to int64) {
	for i, v := range pr {
		fmt.Println("group ", i, "", from, "~", to, ":", v.TimeStamp)
	}
}

func printStatistic(cConf ClusterConfig, stat *GroupsAllStatistic) {
	for i, g := range stat.PingStatisticList {
		statisticTime := fmt.Sprintf("min/mean/max/stddev ms = %d/%3.0f/%d/%3.0f",
			g.TimeStatistic.Min, g.TimeStatistic.Mean, g.TimeStatistic.Max, g.TimeStatistic.Stddev)

		log.Println(fmt.Sprintf("%d->{ hostname: %s, submitted: %3.0f,confirmed:%3.0f, loss: %3.1f%s, count:%d %s}",
			i, cConf.HostName, g.Submitted, g.Confirmed, g.Loss*100, "%", g.Count, statisticTime))
	}
	for i, g := range stat.RawPingStaticList {
		statisticTime := fmt.Sprintf("min/mean/max/stddev ms = %d/%3.0f/%d/%3.0f",
			g.TimeStatistic.Min, g.TimeStatistic.Mean, g.TimeStatistic.Max, g.TimeStatistic.Stddev)
		errString := ""
		for _, v := range g.Errors {
			errString = errString + v + "\n"
		}
		log.Println(fmt.Sprintf("RAW %d->{ hostname: %s, submitted: %3.0f,confirmed:%3.0f, loss: %3.3f%s, count:%d %s}",
			i, cConf.HostName, g.Submitted, g.Confirmed, g.Loss, "%", g.Count, statisticTime))
	}
}
