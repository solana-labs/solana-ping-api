package main

import (
	"encoding/json"
	"os"
	"sync"
)

// Temp database

const MaxRecordInDB = 1000
const HistoryFilepathMainnet = "ping-history-mainnet.serialized"
const HistoryFilepathTestnet = "ping-history-testnet.serialized"
const HistoryFilepathDevnet = "ping-history-devnet.serialized"

type PingHistory []PingResult

var devnetDB PingHistory
var testnetDB PingHistory
var mainnetBetaDB PingHistory

var mutex sync.Mutex

func (s *PingHistory) Add(p PingResult, c Cluster) {
	mutex.Lock()
	*s = append(*s, p)
	mutex.Unlock()
	go SaveToFileHelper(c)
}

func (s *PingHistory) GetLatest(n int) []PingResult {
	// 0 ~ [records-1]
	records := len(*s)
	if records <= 0 {
		return []PingResult{}
	}
	if n > records {
		return (*s)[:]
	}
	return (*s)[records-n:]

}

func (s *PingHistory) GetTimeAfter(ts int64) []PingResult {
	// 0 ~ [records-1]
	records := len(*s)
	if records <= 0 {
		return []PingResult{}
	}
	for i := records - 1; i > 0; i-- {
		if (*s)[i].TimeStamp <= ts {
			if i > 1 {
				log.Info("GetTimeAfter return size=", len((*s)[i:]))
				return (*s)[i:]
			}
		}
	}
	return []PingResult{}
}

func (s *PingHistory) RemoveOld(keep int) int {
	mutex.Lock()
	defer mutex.Unlock()
	records := len(*s)
	if keep < records {
		*s = (*s)[(records - keep):]
		log.Info("remove ", (records - keep), "  new size = ", len(*s))
		return (records - keep)
	}

	return 0
}

func SaveToFileHelper(c Cluster) {
	filepath := ""
	switch c {
	case MainnetBeta:
		filepath = HistoryFilepathMainnet
	case Testnet:
		filepath = HistoryFilepathTestnet
	case Devnet:
		filepath = HistoryFilepathDevnet
	default:
		filepath = HistoryFilepathDevnet
	}
	f, err := os.OpenFile(filepath, os.O_CREATE|os.O_RDWR, 0644)
	defer f.Close()
	if err != nil {
		log.Error("HistoryFile Open Error:", err)
		return
	}
	err = devnetDB.SaveToFile(f)
	if err != nil {
		log.Error("HistoryFile Save Error:", err)
	}
}

func (s *PingHistory) SaveToFile(f *os.File) error {
	data, err := json.Marshal(*s)
	if err != nil {
		log.Error("SaveToFile Error", err)
	}
	_, err = f.Write(data)
	if err != nil {
		return err
	}
	return nil
}

func (s *PingHistory) ReconstructFromFile(f *os.File) error {
	container := make([]byte, 100000)
	n, err := f.Read(container)
	if err != nil {
		return err
	}
	fData := container[:n]
	history := []PingResult{}
	json.Unmarshal(fData, &history)
	*s = history

	return nil
}
