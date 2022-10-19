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
	DataPointReport PingType = "report"
	DataPoint1Min   PingType = "datapoint1min"
)

func launchWorkers(c ClustersToRun) {
	// Run Ping Service
	runCluster := func(clusterConf ClusterConfig) {
		if !clusterConf.PingServiceEnabled {
			log.Println("==> go pingDataWorker", clusterConf.Cluster, " PingServiceEnabled ", clusterConf.PingServiceEnabled)
		} else {
			for i := 0; i < clusterConf.PingConfig.NumWorkers; i++ {
				log.Println("==> go pingDataWorker", clusterConf.Cluster, " n:", clusterConf.PingConfig.NumWorkers, "i:", i)
				go pingDataWorker(clusterConf, i)
				time.Sleep(2 * time.Second)
			}
		}
		if clusterConf.Report.Enabled {
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
		go retensionServiceWorker()
	}
}

func pingDataWorker(cConf ClusterConfig, workerNum int) {
	log.Println(">> Solana DataPoint1MinWorker for ", cConf.Cluster, " worker:", workerNum, " start!")
	defer log.Println(">> Solana DataPoint1MinWorker for ", cConf.Cluster, " worker:", workerNum, " end!")
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
		c = failover.GoNext(c, cConf, workerNum)
		result, err := Ping(c, DataPoint1Min, acct, cConf)
		addRecord(result)
		failover.GetEndpoint().RetryResult(err)
		waitTime := cConf.ClusterPing.PingConfig.MinPerPingTime - (result.TakeTime / 1000)
		if waitTime > 0 {
			time.Sleep(time.Duration(waitTime) * time.Second)
		}
	}
}

func retensionServiceWorker() {
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

func reportWorker(cConf ClusterConfig) {
	log.Println(">> Report Worker for ", cConf.Cluster, " start!")
	defer log.Println(">> Report Worker for ", cConf.Cluster, " end!")
	var lastReporTime int64
	trigger := NewAlertTrigger(cConf)
	for {
		now := time.Now().UTC().Unix()
		if lastReporTime == 0 { // server restart
			lastReporTime = now - int64(cConf.Report.Interval)
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
		trigger.Update(globalStat.Loss)
		alertSend := trigger.ShouldAlertSend() // ShouldAlertSend execute once only. TODO: make shouldAlertSend a function which does not modify any value
		if cConf.Report.Slack.Report.Enabled {
			slackReportSend(cConf, groupsStat, &globalStat)
		}
		if cConf.Report.Slack.Alert.Enabled && alertSend {
			slackAlertSend(cConf, &globalStat, groupsStat.GlobalErrorStatistic, trigger.ThresholdLevels[trigger.ThresholdIndex])
		}
		if cConf.Report.Discord.Report.Enabled {
			discordReportSend(cConf, groupsStat, &globalStat)
		}
		if cConf.Report.Discord.Alert.Enabled && alertSend {
			discordAlertSend(cConf, &globalStat, groupsStat.GlobalErrorStatistic, trigger.ThresholdLevels[trigger.ThresholdIndex])
		}
		time.Sleep(time.Duration(cConf.Report.Interval) * time.Second)
	}
}
func slackReportSend(cConf ClusterConfig, groupsStat *GroupsAllStatistic, globalStat *GlobalStatistic) {
	payload := SlackPayload{}
	payload.ReportPayload(cConf.Cluster, groupsStat, *globalStat)
	err := SlackSend(cConf.Report.Slack.Report.Webhook, &payload)
	if err != nil {
		log.Println("slackReportSend Error:", err)
	}
}
func slackAlertSend(conf ClusterConfig, globalStat *GlobalStatistic, globalErrorStatistic map[string]int, threadhold float64) {
	payload := SlackPayload{}
	payload.AlertPayload(conf, globalStat, globalErrorStatistic, threadhold)
	err := SlackSend(conf.Report.Slack.Alert.Webhook, &payload)
	if err != nil {
		log.Println("slackAlertSend Error:", err)
	}
}

func discordReportSend(cConf ClusterConfig, groupsStat *GroupsAllStatistic, globalStat *GlobalStatistic) {
	payload := DiscordPayload{BotAvatarURL: cConf.Report.Discord.BotAvatarURL, BotName: cConf.Report.Discord.BotName}
	payload.ReportPayload(cConf.Cluster, groupsStat, *globalStat)
	err := DiscordSend(cConf.Report.Discord.Report.Webhook, &payload)
	if err != nil {
		log.Println("discordReportSend Error:", err)
	}
}

func discordAlertSend(cConf ClusterConfig, globalStat *GlobalStatistic, globalErrorStatistic map[string]int, threadhold float64) {
	payload := DiscordPayload{BotAvatarURL: cConf.Report.Discord.BotAvatarURL, BotName: cConf.Report.Discord.BotName}
	payload.AlertPayload(cConf, globalStat, globalErrorStatistic, threadhold)
	err := DiscordSend(cConf.Report.Discord.Alert.Webhook, &payload)
	if err != nil {
		log.Println("discordAlertSend Error:", err)
	}
}
