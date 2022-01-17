package main

import (
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"time"
	"net/http"
	"sync"
)

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

	gormDB, err := gorm.Open(postgres.Open(config.DBConn), &gorm.Config{})
	if err != nil {
		log.Panic(err)
	}
	database = gormDB

	dbMtx = &sync.Mutex{}
	log.Println("database initialized")

}

func main() {
	go launchWorkers(config.Clusters, config.Slack.Clusters)
	router := gin.Default()
	router.GET("/:cluster/latest", getLatest)
	router.GET("/:cluster/last6hours", last6hours)
	router.Run(config.ServerIP)
}

func getLatest(c *gin.Context) {
	cluster := c.Param("cluster")
	var ret PingResultJSON
	switch cluster {
	case "mainnet-beta":
		ret = GetLatestResult(MainnetBeta)
	case "testnet":
		ret = GetLatestResult(Testnet)
	case "devnet":
		ret = GetLatestResult(Devnet)
	default:
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	c.IndentedJSON(http.StatusOK, ret)
}
func last6hours(c *gin.Context) {
	cluster := c.Param("cluster")
	var ret []PingResultJSON
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

//GetLatestResult return the latest PingResult from the cluster and convert it into PingResultJSON
func GetLatestResult(c Cluster) PingResultJSON {
	if !IsClusterActive(c) {
		return PingResultJSON{ErrorMessage: "Cluster " + string(c) + " is not active"}
	}
	records := getLastN(c, 1)
	if len(records) > 0 {
		return ToJoson(&records[0])
	}

	return PingResultJSON{}
}

//GetLatestResult return the latest PingResult from the cluster and convert it into PingResultJSON
func GetLast6hours(c Cluster) []PingResultJSON {
	if !IsClusterActive(c) {
		return []PingResultJSON{}
	}
	now := time.Now().UTC().Unix()
	// (-1) because getAfter function return only after .
	beginOfPast60Hours := now - 60*60*6 - 1 
	records := getAfter(c, beginOfPast60Hours)
	ret := []PingResultJSON{}
	for _, e := range records{
		if len(e.Error) <= 0 { // return only valid data point
			ret = append(ret, ToJoson(&e))
		}
	}
	return ret
}

func IsClusterActive(c Cluster) bool {
	for _, existedCluster := range config.Clusters {
		if c == existedCluster { // cluster existed
			return true
		}
	}
	return false
}
