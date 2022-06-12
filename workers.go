package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"time"

	"github.com/portto/solana-go-sdk/client"
	"github.com/portto/solana-go-sdk/types"
)

type PingType string

const DefaultAlertThredHold = 20
const (
	Report        PingType = "report"
	DataPoint1Min PingType = "datapoint1min"
)

func launchWorkers(c ClustersToRun) {
	// Run API API Service
	// Run Ping Service
	runCluster := func(clusterConf ClusterConfig) {
		if !clusterConf.PingServiceEnabled {
			log.Println("==> go pingDataWorker", clusterConf.Cluster, " PingServiceEnabled ", clusterConf.PingServiceEnabled)
			return
		}
		for i := 0; i < clusterConf.PingConfig.NumWorkers; i++ {
			log.Println("==> go pingDataWorker", clusterConf.Cluster, " n:", clusterConf.PingConfig.NumWorkers, "i:", i)
			go pingDataWorker(clusterConf)
			time.Sleep(2 * time.Second)
		}
		if clusterConf.SlackReport.Enabled {
			go reportWorker(clusterConf)
		}
	}
	// Single Cluster or all Cluster
	switch c {
	case RunMainnetBeta:
		runCluster(config.Mainnet)
	case RunTestnet:
		runCluster(config.Testnet)
	case RunDevnet:
		runCluster(config.Devnet)
	case RunAllClusters:
		runCluster(config.Mainnet)
		runCluster(config.Testnet)
		runCluster(config.Devnet)
	default:
		panic(ErrInvalidCluster)
	}
	// Run Retension Service
	if config.Retension.Enabled {
		time.Sleep(2 * time.Second)
		go RetensionServiceWorker()
	}
}

func pingDataWorker(cConf ClusterConfig) {
	log.Println(">> Solana DataPoint1MinWorker for ", cConf.Cluster, " start!")
	defer log.Println(">> Solana DataPoint1MinWorker for ", cConf.Cluster, " end!")
	var failover RPCFailover
	var c *client.Client
	var acct types.Account

	switch cConf.Cluster {
	case MainnetBeta:
		failover = mainnetFailover
		clusterAcct, err := getConfigKeyPair(config.ClusterCLIConfig.ConfigMain)
		if err != nil {
			log.Panic("getConfigKeyPair Error")
		}
		acct = clusterAcct
	case Testnet:
		failover = testnetFailover
		clusterAcct, err := getConfigKeyPair(config.ClusterCLIConfig.ConfigTestnet)
		if err != nil {
			log.Panic("Testnet getConfigKeyPair Error")
		}
		acct = clusterAcct
	case Devnet:
		failover = devnetFailover
		clusterAcct, err := getConfigKeyPair(config.ClusterCLIConfig.ConfigDevnet)
		if err != nil {
			log.Panic("Devnet getConfigKeyPair Error")
		}
		acct = clusterAcct
	default:
		panic(ErrInvalidCluster)
	}
	for {
		c = failover.GoNext(c, cConf)
		result, err := Ping(c, DataPoint1Min, acct, cConf)
		addRecord(result)
		failover.GetEndpoint().RetryResult(err)
		waitTime := cConf.ClusterPing.PingConfig.MinPerPingTime - (result.TakeTime / 1000)
		if waitTime > 0 {
			time.Sleep(time.Duration(waitTime) * time.Second)
		}
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

func getConfigKeyPair(c SolanaCLIConfig) (types.Account, error) {
	body, err := ioutil.ReadFile(c.KeypairPath)
	if err != nil {
		return types.Account{}, ErrKeyPairFile
	}
	key := []byte{}
	err = json.Unmarshal(body, &key)
	if err != nil {
		return types.Account{}, err
	}

	acct, err := types.AccountFromBytes(key)
	if err != nil {
		return types.Account{}, err
	}
	return acct, nil

}

var lastReporTime int64
var lastReporUnixTime int64

func reportWorker(cConf ClusterConfig) {
	log.Println(">> Slack Report Worker for ", cConf.Cluster, " start!")
	defer log.Println(">> Slack Report Worker for ", cConf.Cluster, " end!")
	slackTrigger := NewAlertTrigger(cConf)
	for {
		now := time.Now().UTC().Unix()
		if lastReporTime == 0 { // server restart
			lastReporTime = now - int64(cConf.SlackReport.ReportInterval)
			log.Println("reconstruct lastReport time=", lastReporTime, "time now=", time.Now().UTC().Unix())
		}
		data := getAfter(cConf.Cluster, DataPoint1Min, lastReporTime)
		if len(data) <= 0 { // No Data
			log.Println(cConf.Cluster, " getAfter return empty")
			time.Sleep(30 * time.Second)
			continue
		}

		groups := grouping1Min(data, lastReporTime, now)
		groupsStat := statisticCompute(cConf, groups)
		globalStat := groupsStat.GetGroupsAllStatistic(false) // get raw data
		lastReporTime = now
		payload := SlackPayload{}
		payload.ReportPayload(cConf.Cluster, groupsStat, globalStat)
		err := SlackSend(cConf.SlackReport.WebHook, &payload)
		if err != nil {
			log.Println("SlackSend Error:", err)
		}

		if cConf.SlackReport.SlackAlert.Enabled {
			slackTrigger.Update(globalStat.Loss)
			if slackTrigger.ShouldAlertSend() {
				AlertSend(cConf, &globalStat, groupsStat.GlobalErrorStatistic, slackTrigger.ThresholdLevels[slackTrigger.ThresholdIndex])
			}

		}
		time.Sleep(time.Duration(cConf.SlackReport.ReportInterval) * time.Second)
	}
}
