package main

import (
	"fmt"
	"time"
)

func PingResultToJson(stat *PingSatistic) DataPoint1MinResultJSON {
	_, mean, _, _, _ := stat.TimeMeasure.Statistic()
	errorShow := ""
	if stat.Count == 0 {
		errorShow = "No Data"
	}
	jsonResult := DataPoint1MinResultJSON{Submitted: int(stat.Submitted), Confirmed: int(stat.Confirmed), Mean: int(mean), Error: errorShow}
	loss := fmt.Sprintf("%3.1f%s", stat.Loss, "%")
	jsonResult.Loss = loss
	ts := time.Unix(stat.TimeStamp, 0).UTC()
	jsonResult.TimeStamp = ts.Format(time.RFC3339)
	return jsonResult
}
