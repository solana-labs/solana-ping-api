package main

import (
	"log"
	"time"
)

type PingType string

const (
	Report        PingType = "report"
	DataPoint1Min PingType = "datapoint1min"
)

func launchWorkers() {
	for _, c := range config.ReportClusters {
		go pingReportWorker(c)
	}
	for _, c := range config.DataPoint1MinClusters {
		go pingDataPoint1MinWorker(c)
	}

	time.Sleep(30 * time.Second)
	for _, c := range config.Slack.Clusters {
		go slackReportWorker(c)
	}

}

func pingReportWorker(c Cluster) {
	log.Println(">> Solana pingReportWorker for ", c, " start!")
	for {
		startTime := time.Now().UTC().Unix()
		result := GetPing(c, Report, config.SolanaPing.Report.Count,
			config.SolanaPing.Report.Interval, int64(config.SolanaPing.Report.Timeout))
		endTime1 := time.Now().UTC().Unix()
		result.PingType = string(Report)
		result.TakeTime = int(endTime1 - startTime)
		addRecord(result)
		endTime2 := time.Now().UTC().Unix()
		waitTime := config.SolanaPing.Report.PerPingTime - (endTime2 - startTime)
		if waitTime > 0 {
			time.Sleep(time.Duration(waitTime) * time.Second)
		}
	}
}

func pingDataPoint1MinWorker(c Cluster) {
	log.Println(">> Solana DataPoint1MinWorker for ", c, " start!")
	for {
		startTime := time.Now().UTC().Unix()
		result := GetPing(c, DataPoint1Min, config.SolanaPing.DataPoint1Min.Count,
			config.SolanaPing.DataPoint1Min.Interval, int64(config.SolanaPing.DataPoint1Min.Timeout))
		endTime1 := time.Now().UTC().Unix()
		result.PingType = string(DataPoint1Min)
		result.TakeTime = int(endTime1 - startTime)
		addRecord(result)
		endTime2 := time.Now().UTC().Unix()
		waitTime := config.SolanaPing.DataPoint1Min.PerPingTime - (endTime2 - startTime)
		if waitTime > 0 {
			time.Sleep(time.Duration(waitTime) * time.Second)
		}
	}
}

//GetPing  Do the solana ping and return ping result, return error is in PingResult.Error
func GetPing(c Cluster, ptype PingType, count int, interval int, timeout int64) PingResult {
	result := PingResult{Hostname: config.HostName, Cluster: string(c)}
	output, err := solanaPing(c, count, interval, timeout)
	if err != nil {
		log.Println(c, " GetPing ping Error:", err)
		result.Error = err.Error()
		if ptype == Report {
			result.Submitted = config.SolanaPing.Report.Count
		} else {
			result.Submitted = config.SolanaPing.DataPoint1Min.Count
		}

		result.Confirmed = 0
		result.Loss = 100
		result.ConfirmationMessage = ""
		result.TimeStamp = time.Now().UTC().Unix()
		return result
	}

	err = result.parsePingOutput(output)
	if err != nil {
		result.Error = err.Error()
		if ptype == Report {
			result.Submitted = config.SolanaPing.Report.Count
		} else {
			result.Submitted = config.SolanaPing.DataPoint1Min.Count
		}
		result.Confirmed = 0
		result.Loss = 100
		result.ConfirmationMessage = ""
		result.TimeStamp = time.Now().UTC().Unix()
		log.Println(c, "GetPing parse output Error:", err, " output:", output)
		return result
	}

	return result
}

var lastReporUnixTime int64

func slackReportWorker(c Cluster) {
	log.Println(">> Slack Report Worker for ", c, " start!")
	for {
		if lastReporUnixTime == 0 {
			lastReporUnixTime = time.Now().UTC().Unix() - int64(config.Slack.ReportTime)
			log.Println("reconstruct lastReport time=", lastReporUnixTime, "time now=", time.Now().UTC().Unix())
		}
		data := getAfter(c, Report, lastReporUnixTime)
		if len(data) <= 0 { // No Data
			log.Println(c, " getAfter return empty")
			time.Sleep(30 * time.Second)
			continue
		}
		lastReporUnixTime = time.Now().UTC().Unix()
		stats := generateReportData(data)
		payload := SlackPayload{}
		payload.ToPayload(c, data, stats)
		err := SlackSend(config.Slack.WebHook, &payload)
		if err != nil {
			log.Println("SlackSend Error:", err)
		}

		time.Sleep(time.Duration(config.Slack.ReportTime) * time.Second)
	}

}
