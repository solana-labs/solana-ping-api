package main

import (
	"fmt"
	"log"
)

type Group1Min struct {
	Result    []PingResult
	TimeStamp int64
}
type TimeStatistic struct {
	Min    int64
	Mean   float64
	Max    int64
	Stddev float64
	Sum    int64
}

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
type GlobalStatistic struct {
	Submitted float64
	Confirmed float64
	Loss      float64
	Count     int
	TimeStatistic
}
type GroupsAllStatistic struct {
	PingStatisticList    []PingSatistic
	RawPingStaticList    []PingSatistic
	GlobalErrorStatistic map[string]int
	ErrorExceptionList   []PingResultError
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

	} else {
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

// setup statistic exception list
func GetStatisticExpections() []PingResultError {
	list := []PingResultError{}
	//list = append(list, RPCServerDeadlineExceeded)
	list = append(list, BlockhashNotFound)
	return list
}

// setup display exception list
func GetDisplayExpections() []PingResultError {
	list := []PingResultError{}
	list = append(list, BlockhashNotFound)
	return list
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

func statisticCompute(groups []Group1Min) *GroupsAllStatistic {
	stat := GroupsAllStatistic{}
	stat.PingStatisticList = []PingSatistic{}
	stat.RawPingStaticList = []PingSatistic{}
	stat.ErrorExceptionList = GetStatisticExpections()
	stat.GlobalErrorStatistic = make(map[string]int)

	for _, group := range groups {
		gStat := PingSatistic{}
		rawGStat := PingSatistic{}
		for _, res := range group.Result {
			errorException := false
			errorCount := len(res.Error)
			if errorCount > 0 {
				for _, e := range res.Error {
					stat.GlobalErrorStatistic[string(e)] = stat.GlobalErrorStatistic[string(e)] + 1
					if PingResultError(e).IsInErrorList(stat.ErrorExceptionList) {
						log.Println("ErrorExceptionList:", e)
						errorException = true
					} else {
						gStat.Errors = append(gStat.Errors, string(e))
					}
					rawGStat.Errors = append(rawGStat.Errors, string(e))
				}
			}
			// Raw Data Statistic
			rawGStat.TimeStamp = group.TimeStamp
			rawGStat.Submitted += float64(res.Submitted)
			rawGStat.Confirmed += float64(res.Confirmed)
			rawGStat.Count += 1
			if errorCount <= 0 {
				rawGStat.TimeMeasure.AddTime(res.TakeTime)
			}
			// Data Statistic (Filtered by error filter)
			gStat.TimeStamp = group.TimeStamp // Need a ts to present the group
			if !errorException {
				gStat.Submitted += float64(res.Submitted)
				gStat.Confirmed += float64(res.Confirmed)
				gStat.Count += 1
				if errorCount <= 0 {
					gStat.TimeMeasure.AddTime(res.TakeTime)
				}
			}

		}
		// raw data
		if rawGStat.Submitted == 0 { // no data
			rawGStat.Loss = 1
		} else {
			rawGStat.Loss = (rawGStat.Submitted - rawGStat.Confirmed) / rawGStat.Submitted
		}
		tMax, tMean, tMin, tStddev, tSum := rawGStat.TimeMeasure.Statistic()
		rawGStat.TimeStatistic = TimeStatistic{
			Min:    tMin,
			Mean:   tMean,
			Max:    tMax,
			Stddev: tStddev,
			Sum:    tSum,
		}
		stat.RawPingStaticList = append(stat.RawPingStaticList, rawGStat)

		// data with filter
		if gStat.Submitted == 0 { // no data
			gStat.Loss = 1
		} else {
			gStat.Loss = (gStat.Submitted - gStat.Confirmed) / gStat.Submitted
		}

		tMax, tMean, tMin, tStddev, tSum = gStat.TimeMeasure.Statistic()
		gStat.TimeStatistic = TimeStatistic{
			Min:    tMin,
			Mean:   tMean,
			Max:    tMax,
			Stddev: tStddev,
			Sum:    tSum,
		}
		stat.PingStatisticList = append(stat.PingStatisticList, gStat)
	}
	return &stat
}

func printPingResultGroup(pr []PingResult, from int64, to int64) {
	for i, v := range pr {
		fmt.Println("group ", i, "", from, "~", to, ":", v.TimeStamp)
	}
}

func PrintStatistic(stat *GroupsAllStatistic) {
	for i, g := range stat.PingStatisticList {

		statisticTime := fmt.Sprintf("min/mean/max/stddev ms = %d/%3.0f/%d/%3.0f",
			g.TimeStatistic.Min, g.TimeStatistic.Mean, g.TimeStatistic.Max, g.TimeStatistic.Stddev)

		log.Println(fmt.Sprintf("%d->{ hostname: %s, submitted: %3.0f,confirmed:%3.0f, loss: %3.1f%s, count:%d %s}",
			i, config.HostName, g.Submitted, g.Confirmed, g.Loss*100, "%", g.Count, statisticTime))
	}
	for i, g := range stat.RawPingStaticList {
		statisticTime := fmt.Sprintf("min/mean/max/stddev ms = %d/%3.0f/%d/%3.0f",
			g.TimeStatistic.Min, g.TimeStatistic.Mean, g.TimeStatistic.Max, g.TimeStatistic.Stddev)
		errString := ""
		for _, v := range g.Errors {
			errString = errString + v + "\n"
		}
		log.Println(fmt.Sprintf("RAW %d->{ hostname: %s, submitted: %3.0f,confirmed:%3.0f, loss: %3.3f%s, count:%d %s}",
			i, config.HostName, g.Submitted, g.Confirmed, g.Loss, "%", g.Count, statisticTime))
	}

	for i, e := range stat.GlobalErrorStatistic {
		log.Println(fmt.Sprintf("{count:%d : %s\n}", e, i))
	}
	for _, e := range stat.ErrorExceptionList {
		log.Println(fmt.Sprintf("ErrorExceptionList: %s", e))
	}
}
