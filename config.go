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
	Clusters           []Cluster
	PingConfig
}

type SlackReport struct {
	Clusters   []Cluster
	WebHook    string
	ReportTime int
	SlackAlert
}
type SlackAlert struct {
	WebHook       string
	LossThredhold int
	LevelFilePath string
}
type ServerSetup struct {
	Mode               ConnectionMode
	IP                 string
	SSLIP              string
	KeyPath            string
	CrtPath            string
	PingService        bool
	RetensionService   bool
	SlackReportService bool
	SlackAlertService  bool
}
type Retension struct {
	KeepHours         int64
	UpdateIntervalSec int64
}

type Config struct {
	UseGCloudDB bool
	ServerSetup
	GCloudCredentialPath string
	DBConn               string
	HostName             string
	ServerIP             string
	Logfile              string
	Tracefile            string
	SolanaConfigInfo
	SolanaPing
	SlackReport
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
	c.ServerSetup.PingService = v.GetBool("ServerSetup.PingService")
	c.ServerSetup.RetensionService = v.GetBool("ServerSetup.RetensionService")
	c.ServerSetup.SlackReportService = v.GetBool("ServerSetup.SlackReportService")
	c.ServerSetup.SlackAlertService = v.GetBool("ServerSetup.SlackAlertService")

	c.UseGCloudDB = v.GetBool("UseGCloudDB")
	c.GCloudCredentialPath = v.GetString("GCloudCredentialPath")
	c.DBConn = v.GetString("DBConn")
	c.HostName = host
	c.ServerIP = v.GetString("ServerIP")
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
	// SolanaPing
	c.SolanaPing = SolanaPing{
		AlternativeEnpoint: EndPoint{
			Mainnet: v.GetString("SolanaPing.AlternativeEnpoint.Mainnet"),
			Testnet: v.GetString("SolanaPing.AlternativeEnpoint.Testnet"),
			Devnet:  v.GetString("SolanaPing.AlternativeEnpoint.Devnet"),
		},
		PingConfig: PingConfig{
			Receiver:                v.GetString("SolanaPing.PingConfig.Receiver"),
			NumWorkers:              v.GetInt("SolanaPing.PingConfig.NumWorkers"),
			BatchCount:              v.GetInt("SolanaPing.PingConfig.BatchCount"),
			BatchInverval:           v.GetInt("SolanaPing.PingConfig.BatchInverval"),
			TxTimeout:               v.GetInt64("SolanaPing.PingConfig.TxTimeout"),
			WaitConfirmationTimeout: v.GetInt64("SolanaPing.PingConfig.WaitConfirmationTimeout"),
			StatusCheckInterval:     v.GetInt64("SolanaPing.PingConfig.StatusCheckInterval"),
			MinPerPingTime:          v.GetInt64("SolanaPing.PingConfig.MinPerPingTime"),
			MaxPerPingTime:          v.GetInt64("SolanaPing.PingConfig.MaxPerPingTime"),
		},
	}
	c.SolanaPing.Clusters = []Cluster{}
	for _, e := range v.GetStringSlice("SolanaPing.Clusters") {
		c.SolanaPing.Clusters = append(c.SolanaPing.Clusters, Cluster(e))
	}
	// SlackReport
	sCluster := []Cluster{}
	for _, e := range v.GetStringSlice("SlackReport.Clusters") {
		sCluster = append(sCluster, Cluster(e))
	}
	c.SlackReport = SlackReport{
		Clusters:   sCluster,
		WebHook:    v.GetString("SlackReport.WebHook"),
		ReportTime: v.GetInt("SlackReport.ReportTime"),
	}
	c.SlackReport.SlackAlert = SlackAlert{
		WebHook:       v.GetString("SlackReport.SlackAlert.WebHook"),
		LossThredhold: v.GetInt("SlackReport.SlackAlert.LossThredhold"),
	}
	levelpath := v.GetString("SlackReport.SlackAlert.LevelFilePath")
	if levelpath != "" {
		c.SlackReport.SlackAlert.LevelFilePath = levelpath
	} else {
		c.SlackReport.SlackAlert.LevelFilePath = "/home/sol/ping-api-server/level.env"
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
