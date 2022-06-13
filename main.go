package main

import (
	"flag"
	"log"
	"strings"
	"sync"
	"time"

	_ "github.com/GoogleCloudPlatform/cloudsql-proxy/proxy/dialers/postgres"
	"github.com/portto/solana-go-sdk/rpc"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var config Config

//Cluster enum
type Cluster string

var database *gorm.DB
var dbMtx *sync.Mutex

const useGCloudDB = true

type ClustersToRun string

//Cluster enum
const (
	MainnetBeta Cluster = "MainnetBeta"
	Testnet             = "Testnet"
	Devnet              = "Devnet"
)

var userInputClusterMode string
var mainnetFailover RPCFailover
var testnetFailover RPCFailover
var devnetFailover RPCFailover

const (
	RunMainnetBeta ClustersToRun = "mainnet"
	RunTestnet                   = "testnet"
	RunDevnet                    = "devnet"
	RunAllClusters               = "all"
)

func init() {
	config = loadConfig()
	log.Println(" *** Config Start *** ")
	log.Println("--- //// Database Config --- ")
	log.Println(config.Database)
	log.Println("--- //// Retension --- ")
	log.Println(config.Retension)
	log.Println("--- //// ClusterCLIConfig--- ")
	log.Println("ClusterCLIConfig Mainnet", config.ClusterCLIConfig.ConfigMain)
	log.Println("ClusterCLIConfig Testnet", config.ClusterCLIConfig.ConfigTestnet)
	log.Println("ClusterCLIConfig Devnet", config.ClusterCLIConfig.ConfigDevnet)
	log.Println("--- Mainnet Ping  --- ")
	log.Println("Mainnet.ClusterPing.APIServer", config.Mainnet.ClusterPing.APIServer)
	log.Println("Mainnet.ClusterPing.PingServiceEnabled", config.Mainnet.ClusterPing.PingServiceEnabled)
	log.Println("Mainnet.ClusterPing.AlternativeEnpoint.HostList", config.Mainnet.ClusterPing.AlternativeEnpoint.HostList)
	log.Println("Mainnet.ClusterPing.PingConfig", config.Mainnet.ClusterPing.PingConfig)
	log.Println("Mainnet.ClusterPing.SlackReport", config.Mainnet.ClusterPing.SlackReport)
	log.Println("--- Testnet Ping  --- ")
	log.Println("Mainnet.ClusterPing.APIServer", config.Testnet.ClusterPing.APIServer)
	log.Println("Mainnet.ClusterPing.PingServiceEnabled", config.Mainnet.ClusterPing.PingServiceEnabled)
	log.Println("Testnet.ClusterPing.AlternativeEnpoint.HostList", config.Testnet.ClusterPing.AlternativeEnpoint.HostList)
	log.Println("Testnet.ClusterPing.PingConfig", config.Testnet.ClusterPing.PingConfig)
	log.Println("Testnet.ClusterPing.SlackReport", config.Testnet.ClusterPing.SlackReport)
	log.Println("--- Devnet Ping  --- ")
	log.Println("Devnet.ClusterPing.APIServer", config.Devnet.ClusterPing.APIServer)
	log.Println("Devnet.ClusterPing.Enabled", config.Devnet.ClusterPing.PingServiceEnabled)
	log.Println("Devnet.ClusterPing.AlternativeEnpoint.HostList", config.Devnet.ClusterPing.AlternativeEnpoint.HostList)
	log.Println("Devnet.ClusterPing.PingConfig", config.Devnet.ClusterPing.PingConfig)
	log.Println("Devnet.ClusterPing.SlackReport", config.Devnet.ClusterPing.SlackReport)
	log.Println(" *** Config End *** ")

	ResponseErrIdentifierInit()
	StatisticErrExpectionInit()
	AlertErrExpectionInit()
	ReportErrExpectionInit()
	PingTakeTimeErrExpectionInit()

	if config.Database.UseGoogleCloud {
		gormDB, err := gorm.Open(postgres.New(postgres.Config{
			DriverName: "cloudsqlpostgres",
			DSN:        config.DBConn,
		}))
		if err != nil {
			log.Panic(err)
		}
		database = gormDB
	} else {
		gormDB, err := gorm.Open(postgres.Open(config.DBConn), &gorm.Config{})
		if err != nil {
			log.Panic(err)
		}
		database = gormDB
	}
	dbMtx = &sync.Mutex{}
	log.Println("database connected")
	/// ---- Start RPC Failover ---
	log.Println("RPC Endpoint Failover Setting ---")
	if len(config.Mainnet.AlternativeEnpoint.HostList) <= 0 {
		mainnetFailover = NewRPCFailover([]RPCEndpoint{{
			Endpoint: rpc.MainnetRPCEndpoint,
			Piority:  1,
			MaxRetry: 30}})
	} else {
		mainnetFailover = NewRPCFailover(config.Mainnet.AlternativeEnpoint.HostList)
	}
	if len(config.Testnet.AlternativeEnpoint.HostList) <= 0 {
		testnetFailover = NewRPCFailover([]RPCEndpoint{{
			Endpoint: rpc.MainnetRPCEndpoint,
			Piority:  1,
			MaxRetry: 30}})
	} else {
		testnetFailover = NewRPCFailover(config.Testnet.AlternativeEnpoint.HostList)
	}
	if len(config.Mainnet.AlternativeEnpoint.HostList) <= 0 {
		devnetFailover = NewRPCFailover([]RPCEndpoint{{
			Endpoint: rpc.MainnetRPCEndpoint,
			Piority:  1,
			MaxRetry: 30}})
	} else {
		devnetFailover = NewRPCFailover(config.Devnet.AlternativeEnpoint.HostList)
	}
}

func main() {
	flag.Parse()
	clustersToRun := flag.Arg(0)
	if !(strings.Compare(clustersToRun, string(RunMainnetBeta)) == 0 ||
		strings.Compare(clustersToRun, string(RunTestnet)) == 0 ||
		strings.Compare(clustersToRun, string(RunDevnet)) == 0 ||
		strings.Compare(clustersToRun, string(RunAllClusters)) == 0) {
		go launchWorkers(RunMainnetBeta)
		go APIService(RunMainnetBeta)
	} else {
		go launchWorkers(ClustersToRun(clustersToRun))
		go APIService(ClustersToRun(clustersToRun))
	}

	for {
		time.Sleep(10 * time.Second)
	}
}
