package main

import (
	"fmt"
	"log"
)

type PingSatistic struct {
	Submitted   float64
	Confirmed   float64
	Loss        float64
	Count       int
	TimeMeasure TakeTime
	Errors      []string
}

type GroupsRawStatistic struct {
	GroupsSatistic     []PingSatistic
	RawGroupsSatistic  []PingSatistic
	ErrList            map[string]int
	ErrorExceptionList []PingResultError
}

// setup statistic exception list
func GetStatisticExpections() []PingResultError {
	list := []PingResultError{}
	list = append(list, RPCServerDeadlineExceeded)
	return list
}

// setup display exception list
func GetDisplayExpections() []PingResultError {
	list := []PingResultError{}
	list = append(list, BlockhashNotFound)
	return list
}

// grouping1Min: group []PingResult into 1 min group.
func grouping1Min(pr []PingResult, startTime int64, endTime int64) [][]PingResult {
	window := int64(60)
	groups := [][]PingResult{}
	for periodend := endTime; periodend > startTime; periodend = periodend - window {
		aGroup := []PingResult{}
		for _, pResult := range pr {
			if pResult.TimeStamp <= periodend && pResult.TimeStamp > periodend-window {
				aGroup = append(aGroup, pResult)
			}
		}
		// printPingResultGroup(aGroup, periodend, periodend-window)
		groups = append(groups, aGroup)
	}
	return groups
}

func statisticCompute(groups [][]PingResult) *GroupsRawStatistic {
	stat := GroupsRawStatistic{}
	stat.GroupsSatistic = []PingSatistic{}
	stat.RawGroupsSatistic = []PingSatistic{}
	stat.ErrorExceptionList = GetStatisticExpections()
	stat.ErrList = make(map[string]int)

	for _, group := range groups {
		gStat := PingSatistic{}
		rawGStat := PingSatistic{}
		for _, res := range group {
			errorException := false
			if len(res.Error) > 0 {
				for _, e := range res.Error {
					stat.ErrList[string(e)] = stat.ErrList[string(e)] + 1
					if PingResultError(e).IsInErrorList(stat.ErrorExceptionList) {
						errorException = true
					}
					gStat.Errors = append(gStat.Errors, string(e))
					rawGStat.Errors = append(rawGStat.Errors, string(e))
				}
			}
			// Raw Data Statistic
			rawGStat.Submitted += float64(res.Submitted)
			rawGStat.Confirmed += float64(res.Confirmed)
			rawGStat.Count += 1
			rawGStat.TimeMeasure.AddTime(res.TakeTime)
			// Filer Data Statistic
			if !errorException {
				gStat.Submitted += float64(res.Submitted)
				gStat.Confirmed += float64(res.Confirmed)
				gStat.Count += 1
				gStat.TimeMeasure.AddTime(res.TakeTime)
			}

		}

		if rawGStat.Submitted == 0 { // no data
			rawGStat.Loss = 1
		} else {
			rawGStat.Loss = (rawGStat.Submitted - rawGStat.Confirmed) / rawGStat.Submitted
		}

		stat.RawGroupsSatistic = append(stat.RawGroupsSatistic, rawGStat)

		if gStat.Submitted == 0 { // no data
			gStat.Loss = 1
		} else {
			gStat.Loss = (gStat.Submitted - gStat.Confirmed) / gStat.Submitted
		}
		stat.GroupsSatistic = append(stat.GroupsSatistic, gStat)
	}
	return &stat
}

func printPingResultGroup(pr []PingResult, from int64, to int64) {
	for i, v := range pr {
		fmt.Println("group ", i, "", from, "~", to, ":", v.TimeStamp)
	}

}

func PrintStatistic(stat *GroupsRawStatistic) {
	for i, g := range stat.GroupsSatistic {
		max, mean, min, stddev, _ := g.TimeMeasure.Statistic()
		statisticTime := fmt.Sprintf("min/mean/max/stddev ms = %d/%3.0f/%d/%3.0f", min, mean, max, stddev)
		errString := ""
		for _, v := range g.Errors {
			errString = errString + v + "\n"
		}
		log.Println(fmt.Sprintf("%d->{ hostname: %s, submitted: %3.0f,confirmed:%3.0f, loss: %3.3f%s, count:%d %s, errors: %s}",
			i, config.HostName, g.Submitted, g.Confirmed, g.Loss, "%", g.Count, statisticTime, errString))
	}
	for i, g := range stat.RawGroupsSatistic {
		max, mean, min, stddev, _ := g.TimeMeasure.Statistic()
		statisticTime := fmt.Sprintf("min/mean/max/stddev ms = %d/%3.0f/%d/%3.0f", min, mean, max, stddev)
		errString := ""
		for _, v := range g.Errors {
			errString = errString + v + "\n"
		}
		log.Println(fmt.Sprintf("RAW %d->{ hostname: %s, submitted: %3.0f,confirmed:%3.0f, loss: %3.3f%s, count:%d %s, errors: %s}",
			i, config.HostName, g.Submitted, g.Confirmed, g.Loss, "%", g.Count, statisticTime, errString))
	}
	for i, e := range stat.ErrList {
		log.Println(fmt.Sprintf("{count:%d : %s\n}", e, i))
	}
	for _, e := range stat.ErrorExceptionList {
		log.Println(fmt.Sprintf("ErrorExceptionList: %s", e))
	}
}
