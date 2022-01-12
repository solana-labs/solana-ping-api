package main

import "time"

const CleanerInterval = 2 * 60 * time.Second

func PingWorkers(clusters []Cluster) {

	for _, c := range clusters {
		go GetPingService(c)
		go CleanerService(c)
	}

}

func GetPingService(c Cluster) {
	for {
		result := GetPing(c)
		devnetDB.Add(result)
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
	for {
		DBCleaner(c)
		time.Sleep(CleanerInterval)
	}
}
func DBCleaner(c Cluster) {
	removed := devnetDB.RemoveOld(MaxRecordInDB)
	log.Info("DBCleaner ", removed, " removed! new : ", devnetDB)
}
