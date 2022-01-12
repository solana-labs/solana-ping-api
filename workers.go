package main

import "time"

const CleanerInterval = 60 * 60 * time.Second

func PingWorkers(clusters []Cluster) {

	for _, c := range clusters {
		go GetPingService(c)
		go CleanerService(c)
	}

}

func GetPingService(c Cluster) {
	for {
		result := GetPing(c)
		devnetDB.Add(result, c)
		time.Sleep(10 * time.Second)
	}
}

func GetPing(c Cluster) PingResult {
	ret := PingResult{Hostname: "aaaa"}
	output, err := solanaPing(Devnet)
	if err != nil {
		ret.ErrorMessage = err
		return ret
	}
	err = ret.parsePingOutput(output)
	if err != nil {
		ret.ErrorMessage = err
		return ret
	}

	return ret
}

func CleanerService(c Cluster) {
	log.Info("Start ", c, "Cleaner Service")
	for {
		switch c {
		case MainnetBeta:
			mainnetBetaDB.RemoveOld(MaxRecordInDB)
		case Testnet:
			testnetDB.RemoveOld(MaxRecordInDB)
		case Devnet:
			devnetDB.RemoveOld(MaxRecordInDB)
		default:
			devnetDB.RemoveOld(MaxRecordInDB)
		}
		time.Sleep(CleanerInterval)
	}
}
