package main

import (
	"log"
	"net/http"
	"os"
	"runtime/trace"
	"sync"
	"time"

	_ "github.com/GoogleCloudPlatform/cloudsql-proxy/proxy/dialers/postgres"
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

//Cluster enum
const (
	MainnetBeta Cluster = "MainnetBeta"
	Testnet             = "Testnet"
	Devnet              = "Devnet"
)

func init() {
	config = loadConfig()

	log.Println("ServerIP:", config.ServerIP, " HostName:", config.HostName,
		" UseGCloudDB:", config.UseGCloudDB, " GCloudCredentialPath", config.GCloudCredentialPath, " DBConn:", config.DBConn, " Logfile:", config.Logfile)
	log.Println("ReportClusters:", config.ReportClusters, " DataPoint1MinClusters:", config.DataPoint1MinClusters)
	log.Println("SolanaConfig/Dir:", config.SolanaConfigInfo.Dir,
		" SolanaConfig/Mainnet", config.SolanaConfigInfo.MainnetPath,
		" SolanaConfig/Testnet", config.SolanaConfigInfo.TestnetPath,
		" SolanaConfig/Devnet", config.SolanaConfigInfo.DevnetPath)
	log.Println("SolanaConfigFile/Mainnet:", config.SolanaConfigInfo.ConfigMain)
	log.Println("SolanaConfigFile/Testnet:", config.SolanaConfigInfo.ConfigTestnet)
	log.Println("SolanaConfigFile/Devnet:", config.SolanaConfigInfo.ConfigDevnet)
	log.Println("SolanaPing:", config.SolanaPing)
	log.Println("Slack:", config.Slack)

	if config.UseGCloudDB {
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

}

func main() {
	f, err := os.Create("/home/sol/wks_go/solana-ping-api/trace.out")
	if err != nil {
		log.Fatalf("failed to create trace output file: %v", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Fatalf("failed to close trace file: %v", err)
		}
	}()

	if err := trace.Start(f); err != nil {
		log.Fatalf("failed to start trace: %v", err)
	}
	defer trace.Stop()

	go launchWorkers()
	router := gin.Default()
	router.GET("/:cluster/latest", getLatest)
	router.GET("/:cluster/last6hours", last6hours)
	router.Run(config.ServerIP)
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
	if !IsReportClusterActive(c) {
		return DataPoint1MinResultJSON{}
	}
	records := getLastN(c, DataPoint1Min, 1)
	if len(records) > 0 {
		return To1MinWindowJson(&records[0])
	}

	return DataPoint1MinResultJSON{}
}

//GetLatestResult return the latest 6hr DataPoint1Min PingResult from the cluster and convert it into PingResultJSON
func GetLast6hours(c Cluster) []DataPoint1MinResultJSON {
	if !IsDataPoint1MinClusterActive(c) {
		return []DataPoint1MinResultJSON{}
	}
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
	results, _ := generateDataPoint1Min(beginOfPast60Hours, now, records)
	return results
}

func IsReportClusterActive(c Cluster) bool {
	for _, existedCluster := range config.ReportClusters {
		if c == existedCluster { // cluster existed
			return true
		}
	}
	return false
}

func IsDataPoint1MinClusterActive(c Cluster) bool {
	for _, existedCluster := range config.DataPoint1MinClusters {
		if c == existedCluster { // cluster existed
			return true
		}
	}
	return false
}
