package main

import (
	"bufio"
	"log"
	"os"
	"strings"

	"github.com/spf13/viper"
)

// ConnectionMode is a type of client connection
type ConnectionMode string

const (
	HTTP  ConnectionMode = "http"
	HTTPS ConnectionMode = "https"
	BOTH  ConnectionMode = "both"
)

type SolanaConfig struct {
	JsonRPCURL    string
	WebsocketURL  string
	KeypairPath   string
	AddressLabels map[string]string
	Commitment    string
}

type SolanaConfigInfo struct {
	Dir           string
	MainnetPath   string
	TestnetPath   string
	DevnetPath    string
	ConfigMain    SolanaConfig
	ConfigTestnet SolanaConfig
	ConfigDevnet  SolanaConfig
}
type PingConfig struct {
	Receiver                string
	NumWorkers              int
	BatchCount              int
	BatchInverval           int
	TxTimeout               int64
	WaitConfirmationTimeout int64
	StatusCheckInterval     int64
	MinPerPingTime          int64
	MaxPerPingTime          int64
}
type EndPoint struct {
	Mainnet string
	Testnet string
	Devnet  string
}
type SolanaPing struct {
	AlternativeEnpoint EndPoint
	Report             PingConfig
	DataPoint1Min      PingConfig
}

type Slack struct {
	Clusters   []Cluster
	WebHook    string
	ReportTime int
}
type ServerSetup struct {
	Mode             ConnectionMode
	IP               string
	SSLIP            string
	KeyPath          string
	CrtPath          string
	NoPingService    bool
	RetensionService bool
}
type Retension struct {
	KeepHours         int64
	UpdateIntervalSec int64
}

type Config struct {
	UseGCloudDB bool
	ServerSetup
	GCloudCredentialPath  string
	DBConn                string
	HostName              string
	ServerIP              string
	Logfile               string
	Tracefile             string
	ReportClusters        []Cluster
	DataPoint1MinClusters []Cluster
	SolanaConfigInfo
	SolanaPing
	Slack
	Retension
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

	c.ServerSetup.Mode = ConnectionMode(v.GetString("ServerSetup.Mode"))

	if c.ServerSetup.Mode != HTTP &&
		c.ServerSetup.Mode != HTTPS && c.ServerSetup.Mode != BOTH {
		c.ServerSetup.Mode = HTTP
		log.Println("server mode not support! use default mode")
	}
	c.ServerSetup.IP = v.GetString("ServerSetup.IP")
	c.ServerSetup.SSLIP = v.GetString("ServerSetup.SSLIP")
	c.ServerSetup.KeyPath = v.GetString("ServerSetup.KeyPath")
	c.ServerSetup.CrtPath = v.GetString("ServerSetup.CrtPath")
	c.ServerSetup.NoPingService = v.GetBool("ServerSetup.NoPingService")
	c.ServerSetup.RetensionService = v.GetBool(("ServerSetup.RetensionService"))
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
	c.Tracefile = v.GetString("Tracefile")
	if len(c.Tracefile) <= 0 {
		c.Tracefile = userHome + "/trace.log"
	}
	c.SolanaConfigInfo = SolanaConfigInfo{
		Dir:         v.GetString("SolanaConfig.Dir"),
		MainnetPath: v.GetString("SolanaConfig.MainnetPath"),
		TestnetPath: v.GetString("SolanaConfig.TestnetPath"),
		DevnetPath:  v.GetString("SolanaConfig.DevnetPath"),
	}
	if len(c.SolanaConfigInfo.MainnetPath) > 0 {
		sConfig, err := ReadSolanaConfigFile(c.SolanaConfigInfo.Dir + c.SolanaConfigInfo.MainnetPath)
		if err != nil {
			log.Fatal(err)
		}
		c.SolanaConfigInfo.ConfigMain = sConfig
	}
	if len(c.SolanaConfigInfo.TestnetPath) > 0 {
		sConfig, err := ReadSolanaConfigFile(c.SolanaConfigInfo.Dir + c.SolanaConfigInfo.TestnetPath)
		if err != nil {
			log.Fatal(err)
		}
		c.SolanaConfigInfo.ConfigTestnet = sConfig
	}
	if len(c.SolanaConfigInfo.DevnetPath) > 0 {
		sConfig, err := ReadSolanaConfigFile(c.SolanaConfigInfo.Dir + c.SolanaConfigInfo.DevnetPath)
		if err != nil {
			log.Fatal(err)
		}
		c.SolanaConfigInfo.ConfigDevnet = sConfig
	}
	c.SolanaPing = SolanaPing{
		AlternativeEnpoint: EndPoint{
			Mainnet: v.GetString("SolanaPing.AlternativeEnpoint.Mainnet"),
			Testnet: v.GetString("SolanaPing.AlternativeEnpoint.Testnet"),
			Devnet:  v.GetString("SolanaPing.AlternativeEnpoint.Devnet"),
		},
		Report: PingConfig{
			Receiver:                v.GetString("SolanaPing.Report.Receiver"),
			NumWorkers:              v.GetInt("SolanaPing.Report.NumWorkers"),
			BatchCount:              v.GetInt("SolanaPing.Report.BatchCount"),
			BatchInverval:           v.GetInt("SolanaPing.Report.BatchInverval"),
			TxTimeout:               v.GetInt64("SolanaPing.Report.TxTimeout"),
			WaitConfirmationTimeout: v.GetInt64("SolanaPing.Report.WaitConfirmationTimeout"),
			StatusCheckInterval:     v.GetInt64("SolanaPing.Report.StatusCheckInterval"),
			MinPerPingTime:          v.GetInt64("SolanaPing.Report.MinPerPingTime"),
			MaxPerPingTime:          v.GetInt64("SolanaPing.Report.MaxPerPingTime"),
		},
		DataPoint1Min: PingConfig{
			Receiver:                v.GetString("SolanaPing.DataPoint1Min.Receiver"),
			NumWorkers:              v.GetInt("SolanaPing.DataPoint1Min.NumWorkers"),
			BatchCount:              v.GetInt("SolanaPing.DataPoint1Min.BatchCount"),
			BatchInverval:           v.GetInt("SolanaPing.DataPoint1Min.BatchInverval"),
			TxTimeout:               v.GetInt64("SolanaPing.DataPoint1Min.TxTimeout"),
			WaitConfirmationTimeout: v.GetInt64("SolanaPing.DataPoint1Min.WaitConfirmationTimeout"),
			StatusCheckInterval:     v.GetInt64("SolanaPing.DataPoint1Min.StatusCheckInterval"),
			MinPerPingTime:          v.GetInt64("SolanaPing.DataPoint1Min.MinPerPingTime"),
			MaxPerPingTime:          v.GetInt64("SolanaPing.DataPoint1Min.MaxPerPingTime"),
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
	gcloudCredential := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	if len(gcloudCredential) == 0 && len(c.GCloudCredentialPath) != 0 {
		os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", c.GCloudCredentialPath)
	}
	c.Retension = Retension{
		KeepHours:         v.GetInt64("Retension.KeepHours"),
		UpdateIntervalSec: v.GetInt64("Retension.UpdateIntervalSec"),
	}
	return c
}

func ReadSolanaConfigFile(filepath string) (SolanaConfig, error) {
	configmap := make(map[string]string, 1)
	addressmap := make(map[string]string, 1)

	f, err := os.Open(filepath)
	if err != nil {
		log.Printf("error opening file: %v\n", err)
		return SolanaConfig{}, err
	}
	r := bufio.NewReader(f)
	line, _, err := r.ReadLine()
	for err == nil {
		k, v := ToKeyPair(string(line))
		if k == "address_labels" {
			line, _, err := r.ReadLine()
			lKey, lVal := ToKeyPair(string(line))
			if err == nil && string(line)[0:1] == " " {
				if len(lKey) > 0 && len(lVal) > 0 {
					addressmap[lKey] = lVal
				}
			} else {
				configmap[k] = v
			}
		} else {
			configmap[k] = v
		}

		line, _, err = r.ReadLine()
	}
	return SolanaConfig{
		JsonRPCURL:    configmap["json_rpc_url"],
		WebsocketURL:  configmap["websocket_url:"],
		KeypairPath:   configmap["keypair_path"],
		AddressLabels: addressmap,
		Commitment:    configmap["commitment"],
	}, nil
}

func ToKeyPair(line string) (key string, val string) {
	noSpaceLine := strings.TrimSpace(string(line))
	idx := strings.Index(noSpaceLine, ":")
	if idx == -1 || idx == 0 { // not found or only have :
		return "", ""
	}
	if (len(noSpaceLine) - 1) == idx { // no value
		return strings.TrimSpace(noSpaceLine[0:idx]), ""
	}
	return strings.TrimSpace(noSpaceLine[0:idx]), strings.TrimSpace(noSpaceLine[idx+1:])
}
