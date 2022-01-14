package main

import (
	"os"

	"github.com/spf13/viper"
)

type SolanaConfig struct {
	Dir     string
	Mainnet string
	Testnet string
	Devnet  string
}
type SolanaPing struct {
	PingExePath string
	Count       int
	Interval    int
	Timeout     int
	PerPingTime int64
}
type Slack struct {
	Clusters   []Cluster
	WebHook    string
	ReportTime int
}
type Cleaner struct {
	MaxRecordInDB   int
	CleanerInterval int
}
type HistoryFile struct {
	Mainnet string
	Testnet string
	Devnet  string
}

type Config struct {
	DBConn    string
	Clusters  []Cluster
	HostName  string
	ServerIP  string
	LogfileOn bool
	Logfile   string
	SolanaConfig
	SolanaPing
	Slack
	Cleaner
	HistoryFile
}

func loadConfig() Config {
	c := Config{}
	v := viper.New()
	v.SetConfigName("config") // 指定 config 的檔名
	v.AddConfigPath("./")
	v.ReadInConfig()
	v.AutomaticEnv()
	host, err := os.Hostname()
	if err != nil {
		c.HostName = ""
	}
	c.DBConn = v.GetString("DBConn")
	c.HostName = host
	c.ServerIP = v.GetString("ServerIP")
	c.Clusters = []Cluster{}
	for _, e := range v.GetStringSlice("Clusters") {
		c.Clusters = append(c.Clusters, Cluster(e))
	}
	c.Logfile = v.GetString("Logfile")
	c.LogfileOn = v.GetBool("LogfileOn")
	c.SolanaConfig = SolanaConfig{
		Dir:     v.GetString("SolanaConfig.Dir"),
		Mainnet: v.GetString("SolanaConfig.Mainnet"),
		Testnet: v.GetString("SolanaConfig.Testnet"),
		Devnet:  v.GetString("SolanaConfig.Devnet"),
	}
	c.SolanaPing = SolanaPing{
		PingExePath: v.GetString("SolanaPing.PingExePath"),
		Count:       v.GetInt("SolanaPing.Count"),
		Interval:    v.GetInt("SolanaPing.Inverval"),
		Timeout:     v.GetInt("SolanaPing.Timeout"),
		PerPingTime: v.GetInt64("SolanaPing.PerPingTime"),
	}
	sCluster := []Cluster{}
	for _, e := range v.GetStringSlice("Slack.Clusters") {
		sCluster = append(sCluster, Cluster(e))
	}
	c.Slack = Slack{
		Clusters:   sCluster,
		WebHook:    v.GetString("Slack.WebHook"),
		ReportTime: v.GetInt("Slack.ReportTime"),
	}
	c.Cleaner = Cleaner{
		MaxRecordInDB:   v.GetInt("Cleaner.MaxRecordInDB"),
		CleanerInterval: v.GetInt("Cleaner.CleanerInterval"),
	}
	c.HistoryFile = HistoryFile{
		Mainnet: v.GetString("History.FilepathMainnet"),
		Testnet: v.GetString("History.FilepathTestnet"),
		Devnet:  v.GetString("History.FilepathDevnet"),
	}

	return c
}
