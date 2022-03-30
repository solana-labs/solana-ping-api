package main

import (
	"github.com/portto/solana-go-sdk/client"
	"github.com/portto/solana-go-sdk/rpc"
	"log"
	"time"
)

type PingType string

const DefaultAlertThredHold = 20
const (
	Report        PingType = "report"
	DataPoint1Min PingType = "datapoint1min"
)

func launchWorkers() {
	if !config.ServerSetup.NoPingService {
		for _, c := range config.ReportClusters {
			for i := 0; i < config.Report.NumWorkers; i++ {
				go pingReportWorker(c)
				time.Sleep(10 * time.Second)
			}
		}
		for _, c := range config.DataPoint1MinClusters {
			for i := 0; i < config.DataPoint1Min.NumWorkers; i++ {
				go pingDataPoint1MinWorker(c)
				time.Sleep(10 * time.Second)
			}

		}
	}

	if !config.ServerSetup.NoSlackReportService {
		time.Sleep(2 * time.Second)
		for _, c := range config.SlackReport.Clusters {
			go slackReportWorker(c)
		}
	}
	if !config.ServerSetup.NoSlackAlertService {
		time.Sleep(2 * time.Second)
		for _, c := range config.SlackAlert.Clusters {
			go slackAlertWorker(c)
		}
	}

	if config.ServerSetup.RetensionService {
		time.Sleep(2 * time.Second)
		go RetensionServiceWorker()
	}

}

func createRPCClient(cluster Cluster) (*client.Client, error) {
	var c *client.Client
	switch cluster {
	case MainnetBeta:
		if len(config.SolanaPing.AlternativeEnpoint.Mainnet) > 0 {
			c = client.NewClient(config.SolanaPing.AlternativeEnpoint.Mainnet)
			log.Println(c, " use alternative endpoint:", config.SolanaPing.AlternativeEnpoint.Mainnet)
		} else {
			c = client.NewClient(rpc.MainnetRPCEndpoint)
		}

	case Testnet:
		if len(config.SolanaPing.AlternativeEnpoint.Testnet) > 0 {
			c = client.NewClient(config.SolanaPing.AlternativeEnpoint.Testnet)
			log.Println(c, " use alternative endpoint:", config.SolanaPing.AlternativeEnpoint.Testnet)
		} else {
			c = client.NewClient(rpc.TestnetRPCEndpoint)
		}
	case Devnet:
		if len(config.SolanaPing.AlternativeEnpoint.Devnet) > 0 {
			c = client.NewClient(config.SolanaPing.AlternativeEnpoint.Devnet)
			log.Println(c, " use alternative endpoint:", config.SolanaPing.AlternativeEnpoint.Devnet)
		} else {
			c = client.NewClient(rpc.DevnetRPCEndpoint)
		}
	default:
		log.Fatal("Invalid Cluster")
		return nil, InvalidCluster
	}
	return c, nil
}

func pingReportWorker(cluster Cluster) {
	log.Println(">> Solana pingReportWorker for ", cluster, " start!")
	defer log.Println(">> Solana pingReportWorker for ", cluster, " end!")
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
	defer log.Println(">> Solana DataPoint1MinWorker for ", cluster, " end!")
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
	defer log.Println(">> Slack Report Worker for ", cluster, " end!")
	for {
		if lastReporUnixTime == 0 {
			lastReporUnixTime = time.Now().UTC().Unix() - int64(config.SlackReport.ReportTime)
			log.Println("reconstruct lastReport time=", lastReporUnixTime, "time now=", time.Now().UTC().Unix())
		}
		data := getAfter(cluster, Report, lastReporUnixTime)
		if len(data) <= 0 { // No Data
			log.Println(cluster, " getAfter return empty")
			time.Sleep(30 * time.Second)
			continue
		}
		lastReporUnixTime = time.Now().UTC().Unix()
		stats := generateStatisticData(data)
		payload := SlackPayload{}
		payload.ToReportPayload(cluster, data, stats)
		err := SlackSend(config.SlackReport.WebHook, &payload)
		if err != nil {
			log.Println("SlackSend Error:", err)
		}

		time.Sleep(time.Duration(config.SlackReport.ReportTime) * time.Second)
	}

}

func slackAlertWorker(cluster Cluster) {
	log.Println(">> Slack Alert Worker for ", cluster, " start!")
	defer log.Println(">> Slack Alert Worker for ", cluster, " end!")
	for {
		lastDataWindowUnixTime := time.Now().UTC().Unix() - int64(config.SlackAlert.DataWindow)
		data := getAfter(cluster, DataPoint1Min, lastDataWindowUnixTime)
		if len(data) <= 0 { // No Data
			log.Println(cluster, " getAfter return empty")
			time.Sleep(30 * time.Second)
			continue
		}
		stats := generateStatisticData(data)
		if config.SlackAlert.LossThredhold < 0 && config.SlackAlert.LossThredhold > 100 { //If not with 0-100 percent range
			config.SlackAlert.LossThredhold = DefaultAlertThredHold
		}
		if stats.Loss*100 > float64(config.SlackAlert.LossThredhold) {
			// send alert
			payload := SlackPayload{}
			payload.ToAlertPayload(cluster, data, stats)
			SlackSend(config.SlackAlert.WebHook, &payload)
		}

		time.Sleep(time.Duration(config.SlackAlert.DataWindow) * time.Second)
	}

}

func RetensionServiceWorker() {
	log.Println(">> Retension Service Worker start!")
	defer log.Println(">> Retension Service Worker end!")
	for {
		now := time.Now().UTC().Unix()
		if config.Retension.KeepHours < 6 {
			config.Retension.KeepHours = 6
		}
		timeB4 := now - (config.Retension.KeepHours * 60 * 60)
		deleteTimeBefore(timeB4)
		if config.Retension.UpdateIntervalSec < 300 {
			config.Retension.UpdateIntervalSec = 300
		}
		time.Sleep(time.Duration(config.Retension.UpdateIntervalSec) * time.Second)
	}
}
