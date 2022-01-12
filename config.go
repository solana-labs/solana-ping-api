package main

import (
	"os"

	"github.com/spf13/viper"
)

type SolanaConfig struct {
	Mainnet string
	Testnet string
	Devnet  string
}
type SolanaPing struct {
	PingExePath string
	Count       int
	Interval    int
	Timeout     int
}
type Slack struct {
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
	HostName         string
	SolanaConfigPath string
	ServerIP         string
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
	c.HostName = host
	c.SolanaConfigPath = v.GetString("SolanaConfigPath")
	c.ServerIP = v.GetString("ServerIP")
	c.SolanaConfig = SolanaConfig{
		Mainnet: v.GetString("SolanaConfig.Mainnet"),
		Testnet: v.GetString("SolanaConfig.Testnet"),
		Devnet:  v.GetString("SolanaConfig.Devnet"),
	}
	c.SolanaPing = SolanaPing{
		PingExePath: v.GetString("SolanaPing.PingExePath"),
		Count:       v.GetInt("SolanaPing.Count"),
		Interval:    v.GetInt("SolanaPing.Inverval"),
		Timeout:     v.GetInt("SolanaPing.Timeout"),
	}
	c.Slack = Slack{
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
