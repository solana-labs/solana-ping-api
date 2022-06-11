package main

import (
	"flag"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	_ "github.com/GoogleCloudPlatform/cloudsql-proxy/proxy/dialers/postgres"
	"github.com/gin-contrib/timeout"
	"github.com/gin-gonic/gin"
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
	RunMainnetBeta ClustersToRun = "MainnetBeta"
	RunTestnet                   = "Testnet"
	RunDevnet                    = "Devnet"
	RunAllClusters               = "All"
)

func init() {
	config = loadConfig()
	log.Println(" *** Config Start *** ")
	log.Println("--- //// Server Config --- ")
	log.Println(config.Server)
	log.Println("--- //// Database Config --- ")
	log.Println(config.Database)
	log.Println("--- //// Retension --- ")
	log.Println(config.Retension)
	log.Println("--- //// ClusterCLIConfig--- ")
	log.Println("ClusterCLIConfig Mainnet", config.ClusterCLIConfig.ConfigMain)
	log.Println("ClusterCLIConfig Testnet", config.ClusterCLIConfig.ConfigTestnet)
	log.Println("ClusterCLIConfig Devnet", config.ClusterCLIConfig.ConfigDevnet)
	log.Println("--- Mainnet Ping  --- ")
	log.Println("Mainnet.ClusterPing.Enabled", config.Mainnet.ClusterPing.Enabled)
	log.Println("Mainnet.ClusterPing.AlternativeEnpoint", config.Mainnet.ClusterPing.AlternativeEnpoint)
	log.Println("Mainnet.ClusterPing.PingConfig", config.Mainnet.ClusterPing.PingConfig)
	log.Println("Mainnet.ClusterPing.SlackReport", config.Mainnet.ClusterPing.SlackReport)
	log.Println("--- Testnet Ping  --- ")
	log.Println("Testnet.ClusterPing.Enabled", config.Testnet.ClusterPing.Enabled)
	log.Println("Testnet.ClusterPing.AlternativeEnpoint", config.Testnet.ClusterPing.AlternativeEnpoint)
	log.Println("Testnet.ClusterPing.PingConfig", config.Testnet.ClusterPing.PingConfig)
	log.Println("Testnet.ClusterPing.SlackReport", config.Testnet.ClusterPing.SlackReport)
	log.Println("--- Devnet Ping  --- ")
	log.Println("Devnet.ClusterPing.Enabled", config.Devnet.ClusterPing.Enabled)
	log.Println("Devnet.ClusterPing.AlternativeEnpoint", config.Devnet.ClusterPing.AlternativeEnpoint)
	log.Println("Devnet.ClusterPing.PingConfig", config.Devnet.ClusterPing.PingConfig)
	log.Println("Devnet.ClusterPing.SlackReport", config.Devnet.ClusterPing.SlackReport)
	log.Println(" *** Config End *** ")

	errList := ResponseErrIdentifierInit()
	log.Println("KnownErrIdentifierInit:", errList)
	errList = StatisticErrExpectionInit()
	log.Println("StatisticErrExpectionInit:", errList)
	errList = AlertErrExpectionInit()
	log.Println("AlertErrExpectionInit:", errList)
	errList = ReportErrExpectionInit()
	log.Println("ReportErrExpectionInit:", errList)
	errList = PingTakeTimeErrExpectionInit()
	log.Println("PingTakeTimeErrExpectionInit:", errList)

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
	/// ---- Start RPC Failover ---
	log.Println("Failover Setting ---")
	mainnetFailover = NewRPCFailover(config.Mainnet.AlternativeEnpoint)
	testnetFailover = NewRPCFailover(config.Testnet.AlternativeEnpoint)
	devnetFailover = NewRPCFailover(config.Devnet.AlternativeEnpoint)

	dbMtx = &sync.Mutex{}
	log.Println("database connected")

}

func main() {
	flag.Parse()
	clustersToRun := flag.Arg(0)
	if !(strings.Compare(clustersToRun, string(RunMainnetBeta)) == 0 ||
		strings.Compare(clustersToRun, string(RunTestnet)) == 0 ||
		strings.Compare(clustersToRun, string(RunDevnet)) == 0) {
		clustersToRun = RunAllClusters
	}

	log.Println("Ping Service will run clusters:", clustersToRun)
	go launchWorkers(ClustersToRun(clustersToRun))

	router := gin.Default()
	router.GET("/:cluster/latest", getLatest)
	router.GET("/:cluster/last6hours", timeout.New(timeout.WithTimeout(10*time.Second), timeout.WithHandler(last6hours)))
	router.GET("/health", health)

	if config.Server.Mode == HTTPS {
		router.RunTLS(config.Server.SSLIP, config.Server.CrtPath, config.Server.KeyPath)
		log.Println("HTTPS server is up!", " IP:", config.Server.SSLIP)
	} else if config.Mode == HTTP {
		log.Println("HTTP server is up!", " IP:", config.Server.IP)
		router.Run(config.Server.IP)
	} else if config.Mode == BOTH {
		go router.RunTLS(config.Server.SSLIP, config.Server.CrtPath, config.Server.KeyPath)
		log.Println("HTTPS server is up!", " IP:", config.Server.SSLIP)
		log.Println("HTTP server is up!", " IP:", config.Server.IP)
		router.Run(config.Server.IP)
	} else {
		log.Panic("Invalid ServerSetup Mode")
	}
}
func health(c *gin.Context) {
	c.Data(200, c.ContentType(), []byte("OK"))
}

func getLatest(c *gin.Context) {
	cluster := c.Param("cluster")
	var ret DataPoint1MinResultJSON
	switch cluster {
	case "mainnet-beta":
		ret = GetLatestResult(MainnetBeta)
	case "testnet":
		ret = GetLatestResult(Testnet)
	case "devnet":
		ret = GetLatestResult(Devnet)
	default:
		c.AbortWithStatus(http.StatusNotFound)
		log.Println("StatusNotFound Error:", cluster)
		return
	}
	c.IndentedJSON(http.StatusOK, ret)
}
func last6hours(c *gin.Context) {
	cluster := c.Param("cluster")
	var ret []DataPoint1MinResultJSON
	switch cluster {
	case "mainnet-beta":
		ret = GetLast6hours(MainnetBeta)
	case "testnet":
		ret = GetLast6hours(Testnet)
	case "devnet":
		ret = GetLast6hours(Devnet)
	default:
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	c.IndentedJSON(http.StatusOK, ret)
}

//GetLatestResult return the latest DataPoint1Min PingResult from the cluster and convert it into PingResultJSON
func GetLatestResult(c Cluster) DataPoint1MinResultJSON {
	records := getLastN(c, DataPoint1Min, 1)
	if len(records) > 0 {
		return To1MinWindowJson(&records[0])
	}

	return DataPoint1MinResultJSON{}
}

//GetLatestResult return the latest 6hr DataPoint1Min PingResult from the cluster and convert it into PingResultJSON
func GetLast6hours(c Cluster) []DataPoint1MinResultJSON {
	lastRecord := getLastN(c, DataPoint1Min, 1)
	now := time.Now().UTC().Unix()
	if len(lastRecord) > 0 {
		now = lastRecord[0].TimeStamp
	}

	// (-1) because getAfter function return only after .
	beginOfPast60Hours := now - 6*60*60
	records := getAfter(c, DataPoint1Min, beginOfPast60Hours)
	if len(records) == 0 {
		return []DataPoint1MinResultJSON{}
	}
	groups := grouping1Min(records, beginOfPast60Hours, now)
	if len(groups) != 6*60 {
		log.Println("WARN! groups is not 360!", " beginOfPast60Hours:", beginOfPast60Hours, "now")
	}

	groupsStat := statisticCompute(GetClusterConfig(c), groups)
	ret := []DataPoint1MinResultJSON{}
	for _, g := range groupsStat.PingStatisticList {
		ret = append(ret, PingResultToJson(&g))
	}
	return ret
}
func GetClusterConfig(c Cluster) ClusterConfig {
	switch c {
	case MainnetBeta:
		return config.Mainnet
	case Testnet:
		return config.Testnet
	case Devnet:
		return config.Devnet
	}
	return ClusterConfig{}
}
