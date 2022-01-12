package main

import (
	"encoding/json"
	"os"
	"sync"
)

// Temp database

const MaxRecordInDB = 3
const HistoryFilepath = "ping-history.serialized"

type PingHistoryDevnet []PingResult

var devnetDB PingHistoryDevnet

var mutex sync.Mutex

func (s *PingHistoryDevnet) Add(p PingResult) {
	mutex.Lock()
	*s = append(*s, p)
	mutex.Unlock()
	go SaveToFileHelper()
}

func (s *PingHistoryDevnet) GetLatest(n int) []PingResult {
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

func (s *PingHistoryDevnet) RemoveOld(keep int) int {
	mutex.Lock()
	defer mutex.Unlock()
	records := len(*s)
	if keep < records {
		*s = (*s)[(records - keep):]
		return (records - keep)
	}

	return 0
}

func SaveToFileHelper() {
	f, err := os.OpenFile(HistoryFilepath, os.O_CREATE|os.O_RDWR, 0644)
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

func (s *PingHistoryDevnet) SaveToFile(f *os.File) error {
	data, err := json.Marshal(*s)
	if err != nil {
		log.Error("SaveToFile Error", err)
	}
	n, err := f.Write(data)
	if err != nil {
		return err
	}
	log.Info("SaveToFile Write :", n, " bytes")
	return nil
}

func (s *PingHistoryDevnet) ReconstructFromFile(f *os.File) error {
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
