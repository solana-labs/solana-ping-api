package main

import (
	"net/http"
	"os"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var log *logrus.Logger
var config Config

//Cluster enum
type Cluster string

var database *gorm.DB
var dbMtx *sync.Mutex

//Cluster enum
const (
	MainnetBeta Cluster = "MainnetBeta"
	Testnet             = "Testnet"
	Devnet              = "Devnet"
)

func init() {
	config = loadConfig()
	log = logrus.New()
	logFormatter := new(logrus.TextFormatter)
	logFormatter.FullTimestamp = true
	log.SetFormatter(logFormatter)

	if config.LogfileOn {
		logfile, err := os.OpenFile(config.Logfile, os.O_WRONLY|os.O_CREATE, 655)
		if err != nil {
			log.Fatal("Log file Error:", err)
			panic("Invalid Cluster")
		}
		log.SetOutput(logfile)
	}

	postgres, err := gorm.Open(postgres.Open(config.DBConn), &gorm.Config{})
	if err != nil {
		log.Panic(err)
	}
	database = postgres
	dbMtx = &sync.Mutex{}
	log.Info("database initialized")

}

func main() {
	go launchWorkers(config.Clusters, config.Slack.Clusters)
	router := gin.Default()
	router.GET("/devnet/latest", getDevnetLatest)
	router.GET("/testnet/latest", getTestnetLatest)
	router.GET("/mainnet-beta/latest", getMainnetBetaLatest)
	router.Run(config.ServerIP)
}

func getMainnetBetaLatest(c *gin.Context) {
	ret := GetLatestResult(MainnetBeta)
	c.IndentedJSON(http.StatusOK, ret)
}

func getTestnetLatest(c *gin.Context) {
	ret := GetLatestResult(Testnet)
	c.IndentedJSON(http.StatusOK, ret)
}

func getDevnetLatest(c *gin.Context) {
	ret := GetLatestResult(Devnet)
	c.IndentedJSON(http.StatusOK, ret)
}

//GetLatestResult return the latest PingResult from the cluster and convert it into PingResultJSON
func GetLatestResult(c Cluster) PingResultJSON {
	if !IsClusterActive(c) {
		return PingResultJSON{ErrorMessage: "Cluster " + string(c) + " is not active"}
	}
	records := getLastN(c, 1)
	if len(records) > 0 {
		return ToJoson(&records[0])
	}

	return PingResultJSON{ErrorMessage: "Invalid Cluster"}
}

func IsClusterActive(c Cluster) bool {
	for _, existedCluster := range config.Clusters {
		if c == existedCluster { // cluster existed
			return true
		}
	}
	return false
}
