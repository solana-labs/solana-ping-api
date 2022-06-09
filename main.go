package main

import (
	"log"
	"net/http"
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

//Cluster enum
const (
	MainnetBeta Cluster = "MainnetBeta"
	Testnet             = "Testnet"
	Devnet              = "Devnet"
)

func init() {
	log.Println("--- Config Start --- ")
	config = loadConfig()
	log.Println("ServerSetup Config:", config.ServerSetup)
	log.Println("Database UseGCloudDB:", config.UseGCloudDB, " GCloudCredentialPath", config.GCloudCredentialPath, " DBConn:", config.DBConn,
		" Logfile:", config.Logfile, " Tracefile:", config.Tracefile)
	log.Println("SolanaConfig/Dir:", config.SolanaConfigInfo.Dir,
		" SolanaConfig/Mainnet", config.SolanaConfigInfo.MainnetPath,
		" SolanaConfig/Testnet", config.SolanaConfigInfo.TestnetPath,
		" SolanaConfig/Devnet", config.SolanaConfigInfo.DevnetPath)
	log.Println("SolanaConfigFile/Mainnet:", config.SolanaConfigInfo.ConfigMain)
	log.Println("SolanaConfigFile/Testnet:", config.SolanaConfigInfo.ConfigTestnet)
	log.Println("SolanaConfigFile/Devnet:", config.SolanaConfigInfo.ConfigDevnet)
	log.Println("SolanaPing:", config.SolanaPing)
	log.Println("SlackReport:", config.SlackReport)
	log.Println("SlackAlert:", config.SlackReport.SlackAlert)
	log.Println("Retension:", config.Retension)
	log.Println("====  Services Setup ===")
	log.Println("PingService:", config.ServerSetup.PingService)
	log.Println("RetensionService:", config.ServerSetup.RetensionService)
	log.Println("SlackReportService:", config.ServerSetup.SlackReportService)
	log.Println("SlackAlertService:", config.ServerSetup.SlackAlertService)
	log.Println("--- Config End --- ")
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
	// f, err := os.Create(config.Tracefile)
	// if err != nil {
	// 	log.Fatalf("failed to create trace output file: %v", err)
	// }
	// defer func() {
	// 	if err := f.Close(); err != nil {
	// 		log.Fatalf("failed to close trace file: %v", err)
	// 	}
	// }()

	// if err := trace.Start(f); err != nil {
	// 	log.Fatalf("failed to start trace: %v", err)
	// }
	// defer trace.Stop()
	go launchWorkers()

	router := gin.Default()
	router.GET("/:cluster/latest", getLatest)
	router.GET("/:cluster/last6hours", timeout.New(timeout.WithTimeout(10*time.Second), timeout.WithHandler(last6hours)))
	router.GET("/health", health)
	router.GET("/test", test)

	if config.ServerSetup.Mode == HTTPS {
		router.RunTLS(config.ServerSetup.SSLIP, config.ServerSetup.CrtPath, config.ServerSetup.KeyPath)
		log.Println("HTTPS server is up!", " IP:", config.ServerSetup.SSLIP)
	} else if config.Mode == HTTP {
		log.Println("HTTP server is up!", " IP:", config.ServerSetup.IP)
		router.Run(config.ServerSetup.IP)
	} else if config.Mode == BOTH {
		go router.RunTLS(config.ServerSetup.SSLIP, config.ServerSetup.CrtPath, config.ServerSetup.KeyPath)
		log.Println("HTTPS server is up!", " IP:", config.ServerSetup.SSLIP)
		log.Println("HTTP server is up!", " IP:", config.ServerSetup.IP)
		router.Run(config.ServerSetup.IP)
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
	if !IsClusterActive(c) && config.ServerSetup.PingService {
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
	if !IsClusterActive(c) && config.ServerSetup.PingService {
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
	groups := grouping1Min(records, beginOfPast60Hours, now)
	if len(groups) != 6*60 {
		log.Println("WARN! groups is not 360!", " beginOfPast60Hours:", beginOfPast60Hours, "now")
	}
	groupsStat := statisticCompute(groups)
	ret := []DataPoint1MinResultJSON{}
	for _, g := range groupsStat.PingStatisticList {
		ret = append(ret, PingResultToJson(&g))
	}
	return ret
}

func IsClusterActive(c Cluster) bool {
	for _, existedCluster := range config.SolanaPing.Clusters {
		if c == existedCluster { // cluster existed
			return true
		}
	}
	return false
}

func test(c *gin.Context) {
	now := time.Now().UTC().Unix()
	//now := int64(1648733636)
	beginOfPast10min := now - 10*60
	records := getAfter(MainnetBeta, DataPoint1Min, beginOfPast10min)
	groups := grouping1Min(records, beginOfPast10min, now)
	groupsStat := statisticCompute(groups)
	printStatistic(groupsStat)
	ret := []DataPoint1MinResultJSON{}
	for _, g := range groupsStat.PingStatisticList {
		ret = append(ret, PingResultToJson(&g))
	}
	c.IndentedJSON(http.StatusOK, ret)
}
