package main

import "time"

func PingWorkers(clusters []Cluster) {

	for _, c := range clusters {
		go GetPingService(c)
		go CleanerService(c)
	}

}

func GetPingService(c Cluster) {
	for {
		result := GetPing(c)

		switch c {
		case MainnetBeta:
			mainnetBetaDB.Add(result, c)
		case Testnet:
			testnetDB.Add(result, c)
		case Devnet:
			devnetDB.Add(result, c)
		default:
			devnetDB.Add(result, c)
		}
		time.Sleep(10 * time.Second)
	}
}

func GetPing(c Cluster) PingResult {
	ret := PingResult{Hostname: config.HostName, Cluster: c}
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
			mainnetBetaDB.RemoveOld(config.Cleaner.MaxRecordInDB)
		case Testnet:
			testnetDB.RemoveOld(config.Cleaner.MaxRecordInDB)
		case Devnet:
			devnetDB.RemoveOld(config.Cleaner.MaxRecordInDB)
		default:
			devnetDB.RemoveOld(config.Cleaner.MaxRecordInDB)
		}
		time.Sleep(time.Duration(config.Cleaner.CleanerInterval) * time.Second)
	}
}
