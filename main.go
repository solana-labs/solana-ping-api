package main

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

func getDevnetLatest(c *gin.Context) {
	ret := GetDevnetLatest()
	c.IndentedJSON(http.StatusOK, ret)
}

func init() {
	ConfigPath = os.Getenv("HOME") + "/.config/solana/cli/"
	log = logrus.New()
	devnetDB = make([]PingResult, 0)
	df, err := os.Open(HistoryFilepathMainnet)
	if nil == err {
		devnetDB.ReconstructFromFile(df)
		defer df.Close()
	}

	testnetDB = make([]PingResult, 0)
	tf, err := os.Open(HistoryFilepathTestnet)
	if nil == err {
		devnetDB.ReconstructFromFile(tf)
		defer tf.Close()
	}
	mainnetBetaDB = make([]PingResult, 0)
	mf, err := os.Open(HistoryFilepathDevnet)
	if nil == err {
		devnetDB.ReconstructFromFile(mf)
		defer mf.Close()
	}
}

func main() {
	router := gin.Default()
	router.GET("/devnet/latest", getDevnetLatest)
	go PingWorkers([]Cluster{Devnet})
	go SlackReportService()
	router.Run("localhost:8080")
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
