package main

import (
	"log"
	"time"
)

func launchWorkers(clusters []Cluster, slackCluster []Cluster) {
	for _, c := range clusters {
		go getPingWorker(c)
	}

	time.Sleep(30 * time.Second)

	for _, c := range slackCluster {
		go slackReportWorker(c)
	}

}

func getPingWorker(c Cluster) {
	log.Println(">> Solana Ping Worker for ", c, " start!")
	for {
		startTime := time.Now().UTC().Unix()
		result := GetPing(c)
		endTime1 := time.Now().UTC().Unix()
		result.TakeTime = int(endTime1 - startTime)
		addRecord(result)
		endTime2 := time.Now().UTC().Unix()
		perPingTime := config.SolanaPing.PerPingTime
		waitTime := perPingTime - (endTime2 - startTime)
		if waitTime > 0 {
			time.Sleep(time.Duration(waitTime) * time.Second)
		}
	}
}

//GetPing  Do the solana ping and return ping result, return error is in PingResult.Error
func GetPing(c Cluster) PingResult {
	result := PingResult{Hostname: config.HostName, Cluster: string(c)}
	output, err := solanaPing(c)
	if err != nil {
		log.Println(c, " GetPing ping Error:", err)
		result.Error = err.Error()
		result.Submitted = config.Count
		result.Confirmed = 0
		result.Loss = 100
		result.ConfirmationMessage = ""
		result.TimeStamp = time.Now().UTC().Unix()
		return result
	}

	err = result.parsePingOutput(output)
	if err != nil {
		result.Error = err.Error()
		result.Submitted = config.Count
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
		data := getAfter(c, lastReporUnixTime)
		if len(data) <= 0 { // No Data
			log.Println(c, " getAfter return empty")
			time.Sleep(30 * time.Second)
			continue
		}
		lastReporUnixTime = time.Now().UTC().Unix()
		stats := generateData(data)
		payload := SlackPayload{}
		payload.ToPayload(c, data, stats)
		err := SlackSend(config.Slack.WebHook, &payload)
		if err != nil {
			log.Println("SlackSend Error:", err)
		}

		time.Sleep(time.Duration(config.Slack.ReportTime) * time.Second)
	}

}
