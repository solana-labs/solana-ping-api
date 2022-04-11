package main

import (
	"log"
	"time"

	"github.com/portto/solana-go-sdk/client"
	"github.com/portto/solana-go-sdk/rpc"
)

type PingType string

const DefaultAlertThredHold = 20
const (
	Report        PingType = "report"
	DataPoint1Min PingType = "datapoint1min"
)

func launchWorkers() {
	if config.ServerSetup.PingService {
		for _, c := range config.SolanaPing.Clusters {

			for i := 0; i < config.PingConfig.NumWorkers; i++ {
				go pingDataWorker(c)
				time.Sleep(10 * time.Second)
			}

		}
	}

	if config.ServerSetup.RetensionService {
		time.Sleep(2 * time.Second)
		go RetensionServiceWorker()
	}

	if config.ServerSetup.SlackReportService {
		for _, c := range config.SlackReport.Clusters {
			go reportWorker(c)
		}
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

func pingDataWorker(cluster Cluster) {
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
		result, err := Ping(cluster, c, config.HostName, DataPoint1Min, config.SolanaPing.PingConfig)
		if err != nil {
			log.Println("pingReportWorker Error:", err)
			continue
		}
		addRecord(result)
		waitTime := config.SolanaPing.PingConfig.MinPerPingTime - (result.TakeTime / 1000)
		if waitTime > 0 {
			//log.Println("---wait for ---", waitTime, " sec")
			time.Sleep(time.Duration(waitTime) * time.Second)
		}
	}
}

var lastReporUnixTime int64

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

var lastReporTime int64

func reportWorker(cluster Cluster) {
	log.Println(">> Slack Report Worker for ", cluster, " start!")
	defer log.Println(">> Slack Report Worker for ", cluster, " end!")
	slackTrigger := NewSlackTriggerEvaluation()
	for {
		now := time.Now().UTC().Unix()
		if lastReporTime == 0 { // server restart
			lastReporTime = now - int64(config.SlackReport.ReportTime)
			log.Println("reconstruct lastReport time=", lastReporTime, "time now=", time.Now().UTC().Unix())
		}
		data := getAfter(cluster, DataPoint1Min, lastReporTime)
		if len(data) <= 0 { // No Data
			log.Println(cluster, " getAfter return empty")
			time.Sleep(30 * time.Second)
			continue
		}

		groups := grouping1Min(data, lastReporTime, now)
		groupsStat := statisticCompute(groups)
		globalStat := groupsStat.GetGroupsAllStatistic(false) // get raw data
		//PrintStatistic(groupsStat)
		lastReporTime = now
		payload := SlackPayload{}
		payload.ReportPayload(cluster, groupsStat, globalStat)
		err := SlackSend(config.SlackReport.WebHook, &payload)
		if err != nil {
			log.Println("SlackSend Error:", err)
		}
		if config.ServerSetup.SlackAlertService {
			slackTrigger.Update(globalStat.Loss)
			log.Println("**slackTrigger.Update:", slackTrigger)
			if slackTrigger.ShouldSend() {
				slackSend(cluster, &globalStat, groupsStat.GlobalErrorStatistic, slackTrigger.ThresHoldLevels[slackTrigger.ThresHoldIndex])
			}

		}
		time.Sleep(time.Duration(config.SlackReport.ReportTime) * time.Second)
	}
}
