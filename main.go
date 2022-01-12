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
	log = logrus.New()
	devnetDB = make([]PingResult, 0)
	f, err := os.Open(HistoryFilepath)
	if nil == err {
		devnetDB.ReconstructFromFile(f)
		defer f.Close()
	}
}

func main() {
	router := gin.Default()
	router.GET("/devnet/latest", getDevnetLatest)
	go PingWorkers([]Cluster{Devnet})
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
