package main

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

var log *logrus.Logger
var config Config

type Cluster string

const (
	MainnetBeta Cluster = "mainnet-beta"
	Testnet             = "testnet"
	Devnet              = "devnet"
)

func init() {
	config = loadConfig()
	log = logrus.New()
	devnetDB = make([]PingResult, 0)
	df, err := os.Open(config.HistoryFile.Devnet)
	if nil != err {
		log.Warn("open devent file fail:", err)
	}
	devnetDB.ReconstructFromFile(df)
	log.Info("devnetDB is recontruct from ", config.HistoryFile.Devnet)
	defer df.Close()

	testnetDB = make([]PingResult, 0)
	tf, err := os.Open(config.HistoryFile.Testnet)
	if nil != err {
		log.Warn("open testnet file fail:", err)
	}
	testnetDB.ReconstructFromFile(tf)
	log.Info("testnetDB is recontruct from ", config.HistoryFile.Testnet)
	defer tf.Close()

	mainnetBetaDB = make([]PingResult, 0)
	mf, err := os.Open(config.HistoryFile.Mainnet)
	if nil != err {
		log.Warn("open mainnet file fail:", err)
	}
	mainnetBetaDB.ReconstructFromFile(mf)
	log.Info("mainnetBetaDB is recontruct from ", config.HistoryFile.Mainnet)
	defer mf.Close()
}

func main() {
	router := gin.Default()
	router.GET("/devnet/latest", getDevnetLatest)
	go PingWorkers([]Cluster{Testnet, Devnet, MainnetBeta})
	go SlackReportService()
	router.Run(config.ServerIP)
}

func getDevnetLatest(c *gin.Context) {
	ret := GetDevnetLatest()
	c.IndentedJSON(http.StatusOK, ret)
}

func GetDevnetLatest() PingResultJson {
	r := devnetDB.GetLatest(1)

	if len(r) > 0 {
		ret, err := r[0].ConvertToJoson()
		if err != nil {
			return PingResultJson{ErrorMessage: err.Error()}
		}
		return ret
	}
	return PingResultJson{ErrorMessage: NoPingResultFound.Error()}
}

func GetTestnetLatest() PingResultJson {
	r := testnetDB.GetLatest(1)

	if len(r) > 0 {
		ret, err := r[0].ConvertToJoson()
		if err != nil {
			return PingResultJson{ErrorMessage: err.Error()}
		}
		return ret
	}
	return PingResultJson{ErrorMessage: NoPingResultFound.Error()}
}

func GetMainnetLatest() PingResultJson {
	r := mainnetBetaDB.GetLatest(1)

	if len(r) > 0 {
		ret, err := r[0].ConvertToJoson()
		if err != nil {
			return PingResultJson{ErrorMessage: err.Error()}
		}
		return ret
	}
	return PingResultJson{ErrorMessage: NoPingResultFound.Error()}
}
