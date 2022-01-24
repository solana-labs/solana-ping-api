package main

import (
	"log"
	"time"

	"github.com/portto/solana-go-sdk/client"
	"github.com/portto/solana-go-sdk/rpc"
)

type PingType string

const (
	Report        PingType = "report"
	DataPoint1Min PingType = "datapoint1min"
)

func launchWorkers() {
	for _, c := range config.ReportClusters {
		for i := 0; i < config.Report.NumWorkers; i++ {
			go pingReportWorker(c)
			time.Sleep(2 * time.Second)
		}
	}
	for _, c := range config.DataPoint1MinClusters {
		for i := 0; i < config.DataPoint1Min.NumWorkers; i++ {
			go pingDataPoint1MinWorker(c)
			time.Sleep(2 * time.Second)
		}

	}

	time.Sleep(30 * time.Second)
	for _, c := range config.Slack.Clusters {
		go slackReportWorker(c)
	}

}

func createRPCClient(cluster Cluster) (*client.Client, error) {
	var c *client.Client
	switch cluster {
	case MainnetBeta:
		c = client.NewClient(rpc.MainnetRPCEndpoint)
	case Testnet:
		c = client.NewClient(rpc.TestnetRPCEndpoint)
	case Devnet:
		c = client.NewClient(rpc.DevnetRPCEndpoint)
	default:
		log.Fatal("Invalid Cluster")
		return nil, InvalidCluster
	}
	return c, nil
}

func pingReportWorker(cluster Cluster) {
	log.Println(">> Solana pingReportWorker for ", cluster, " start!")
	c, err := createRPCClient(cluster)
	if err != nil {
		return
	}
	for {
		if c == nil {
			c, err = createRPCClient(cluster)
			if err != nil {
				return
			}
		}
		result, err := Ping(cluster, c, config.HostName, Report, config.SolanaPing.Report)
		if err != nil {
			log.Println("pingReportWorker Error:", err)
			continue
		}
		addRecord(result)
		waitTime := config.SolanaPing.Report.MinPerPingTime - result.TakeTime
		if waitTime > 0 {
			time.Sleep(time.Duration(waitTime) * time.Second)
		}
	}
}

func pingDataPoint1MinWorker(cluster Cluster) {
	log.Println(">> Solana DataPoint1MinWorker for ", cluster, " start!")
	c, err := createRPCClient(cluster)
	if err != nil {
		return
	}
	for {
		if c == nil {
			c, err = createRPCClient(cluster)
			if err != nil {
				return
			}
		}
		result, err := Ping(cluster, c, config.HostName, DataPoint1Min, config.SolanaPing.DataPoint1Min)
		if err != nil {
			log.Println("pingReportWorker Error:", err)
			continue
		}
		addRecord(result)
		waitTime := config.SolanaPing.DataPoint1Min.MinPerPingTime - (result.TakeTime / 1000)
		if waitTime > 0 {
			log.Println("---wait for ---", waitTime, " sec")
			time.Sleep(time.Duration(waitTime) * time.Second)
		}
	}
}

var lastReporUnixTime int64

func slackReportWorker(cluster Cluster) {
	log.Println(">> Slack Report Worker for ", cluster, " start!")
	for {
		if lastReporUnixTime == 0 {
			lastReporUnixTime = time.Now().UTC().Unix() - int64(config.Slack.ReportTime)
			log.Println("reconstruct lastReport time=", lastReporUnixTime, "time now=", time.Now().UTC().Unix())
		}
		data := getAfter(cluster, Report, lastReporUnixTime)
		if len(data) <= 0 { // No Data
			log.Println(cluster, " getAfter return empty")
			time.Sleep(30 * time.Second)
			continue
		}
		lastReporUnixTime = time.Now().UTC().Unix()
		stats := generateReportData(data)
		payload := SlackPayload{}
		payload.ToPayload(cluster, data, stats)
		err := SlackSend(config.Slack.WebHook, &payload)
		if err != nil {
			log.Println("SlackSend Error:", err)
		}

		time.Sleep(time.Duration(config.Slack.ReportTime) * time.Second)
	}

}
