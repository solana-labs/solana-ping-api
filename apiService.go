package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-contrib/timeout"
	"github.com/gin-gonic/gin"
)

func APIService(c ClustersToRun) {
	runCluster := func(mode ConnectionMode, host string, hostSSL string, key string, crt string) {
		router := gin.Default()
		router.GET("/:cluster/latest", getLatest)
		router.GET("/:cluster/last6hours", timeout.New(timeout.WithTimeout(10*time.Second), timeout.WithHandler(last6hours)))
		router.GET("/health", health)
		if mode == HTTPS {
			err := router.RunTLS(hostSSL, crt, key)
			if err != nil {
				log.Panic("api service is not up!!!", err)
				panic("api service is not up!!!")
			}
			log.Println("HTTPS server is up!", " Server:", host)
		} else if mode == HTTP {
			err := router.Run(host)
			if err != nil {
				log.Panic("api service is not up!!!", err)
				panic("api service is not up!!!")
			}
			log.Println("HTTP server is up!", " Server", host)
		} else if config.Mainnet.APIServer.Mode == BOTH {
			err := router.RunTLS(host, crt, key)
			if err != nil {
				log.Panic("api service is not up!!!", err)
				panic("api service is not up!!!")
			}
			log.Println("HTTPS server is up!", " Server:", host)
			err = router.Run(host)
			log.Println("HTTP server is up!", " Server", host)
			if err != nil {
				log.Panic("api service is not up!!!", err)
				panic("api service is not up!!!")
			}
		} else {
			log.Panic("Invalid ServerSetup Mode")
		}
	}
	// Single Cluster or all Cluster
	switch c {
	case RunMainnetBeta:
		if config.Mainnet.APIServer.Enabled {
			go runCluster(config.Mainnet.APIServer.Mode,
				config.Mainnet.APIServer.IP,
				config.Mainnet.APIServer.SSLIP,
				config.Mainnet.APIServer.KeyPath,
				config.Mainnet.APIServer.CrtPath)
			log.Println("--- API Server Mainnet Start--- ")
		}

	case RunTestnet:
		if config.Testnet.APIServer.Enabled {
			go runCluster(config.Testnet.APIServer.Mode,
				config.Testnet.APIServer.IP,
				config.Testnet.APIServer.SSLIP,
				config.Testnet.APIServer.KeyPath,
				config.Testnet.APIServer.CrtPath)
			log.Println("--- API Server Testnet Start--- ")
		}

	case RunDevnet:
		if config.Devnet.APIServer.Enabled {
			go runCluster(config.Devnet.APIServer.Mode,
				config.Devnet.APIServer.IP,
				config.Devnet.APIServer.SSLIP,
				config.Devnet.APIServer.KeyPath,
				config.Devnet.APIServer.CrtPath)
			log.Println("--- API Server Devnet Start--- ")
		}
	case RunAllClusters:
		if config.Mainnet.APIServer.Enabled {
			go runCluster(config.Mainnet.APIServer.Mode,
				config.Mainnet.APIServer.IP,
				config.Mainnet.APIServer.SSLIP,
				config.Mainnet.APIServer.KeyPath,
				config.Mainnet.APIServer.CrtPath)
			log.Println("--- API Server Mainnet Start--- ")
		}
		if config.Testnet.APIServer.Enabled {
			go runCluster(config.Testnet.APIServer.Mode,
				config.Testnet.APIServer.IP,
				config.Testnet.APIServer.SSLIP,
				config.Testnet.APIServer.KeyPath,
				config.Testnet.APIServer.CrtPath)
			log.Println("--- API Server Testnet Start--- ")
		}
		if config.Devnet.APIServer.Enabled {
			go runCluster(config.Devnet.APIServer.Mode,
				config.Devnet.APIServer.IP,
				config.Devnet.APIServer.SSLIP,
				config.Devnet.APIServer.KeyPath,
				config.Devnet.APIServer.CrtPath)
			log.Println("--- API Server Devnet Start--- ")
		}

	default:
		panic(ErrInvalidCluster)
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
