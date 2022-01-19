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
type PingConfig struct {
	Count       int
	Interval    int
	Timeout     int
	PerPingTime int64
}
type SolanaPing struct {
	PingExePath   string
	Report        PingConfig
	DataPoint1Min PingConfig
}
type Slack struct {
	Clusters   []Cluster
	WebHook    string
	ReportTime int
}

type Config struct {
	UseGCloudDB           bool
	GCloudCredentialPath  string
	DBConn                string
	HostName              string
	ServerIP              string
	Logfile               string
	ReportClusters        []Cluster
	DataPoint1MinClusters []Cluster
	SolanaConfig
	SolanaPing
	Slack
}

func loadConfig() Config {
	userHome, err := os.UserHomeDir()
	if err != nil {
		panic("loadConfig error:" + err.Error())
	}
	c := Config{}
	v := viper.New()
	v.SetConfigName("config")
	v.AddConfigPath(userHome + "/.config/ping-api")
	v.ReadInConfig()
	v.AutomaticEnv()
	host, err := os.Hostname()
	if err != nil {
		c.HostName = ""
	}
	c.UseGCloudDB = v.GetBool("UseGCloudDB")
	c.GCloudCredentialPath = v.GetString("GCloudCredentialPath")
	c.DBConn = v.GetString("DBConn")
	c.HostName = host
	c.ServerIP = v.GetString("ServerIP")

	c.ReportClusters = []Cluster{}
	for _, e := range v.GetStringSlice("Clusters.Report") {
		c.ReportClusters = append(c.ReportClusters, Cluster(e))
	}
	c.DataPoint1MinClusters = []Cluster{}
	for _, e := range v.GetStringSlice("Clusters.DataPoint1Min") {
		c.DataPoint1MinClusters = append(c.DataPoint1MinClusters, Cluster(e))
	}
	c.Logfile = v.GetString("Logfile")
	c.SolanaConfig = SolanaConfig{
		Dir:     v.GetString("SolanaConfig.Dir"),
		Mainnet: v.GetString("SolanaConfig.Mainnet"),
		Testnet: v.GetString("SolanaConfig.Testnet"),
		Devnet:  v.GetString("SolanaConfig.Devnet"),
	}
	c.SolanaPing = SolanaPing{
		PingExePath: v.GetString("SolanaPing.PingExePath"),
		Report: PingConfig{
			Count:       v.GetInt("SolanaPing.Report.Count"),
			Interval:    v.GetInt("SolanaPing.Report.Inverval"),
			Timeout:     v.GetInt("SolanaPing.Report.Timeout"),
			PerPingTime: v.GetInt64("SolanaPing.Report.PerPingTime"),
		},
		DataPoint1Min: PingConfig{
			Count:       v.GetInt("SolanaPing.DataPoint1Min.Count"),
			Interval:    v.GetInt("SolanaPing.DataPoint1Min.Inverval"),
			Timeout:     v.GetInt("SolanaPing.DataPoint1Min.Timeout"),
			PerPingTime: v.GetInt64("SolanaPing.DataPoint1Min.PerPingTime"),
		},
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
	osPath := os.Getenv("PATH")
	if len(osPath) != 0 {
		osPath = c.PingExePath + ":" + osPath
		os.Setenv("PATH", osPath)
	}
	os.Setenv("PATH", c.PingExePath)
	gcloudCredential := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	if 0 == len(gcloudCredential) {
		os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", config.GCloudCredentialPath)
	}

	return c
}
