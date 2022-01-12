package main

import (
	"errors"
	"os"
	"testing"
)

var pingResult1 = PingResult{Hostname: "a", Submitted: 10, Confirmed: 10, Loss: 0, ConfirmationMessage: "nothing", ErrorMessage: nil}
var pingResult2 = PingResult{Hostname: "b", Submitted: 10, Confirmed: 0, Loss: 100, ConfirmationMessage: "nothing", ErrorMessage: nil}
var pingResult3 = PingResult{Hostname: "c", Submitted: 5, Confirmed: 5, Loss: 50, ConfirmationMessage: "nothing", ErrorMessage: errors.New("OooMyErr")}

/*
func TestSaveToFile(t *testing.T) {
	os.Remove(HistoryFilepath)
	devnetDB = make([]PingResult, 0)
	devnetDB.Add(pingResult1)
	devnetDB.Add(pingResult2)
	devnetDB.Add(pingResult3)
	SaveToFileHelper()

}

func TestOpenHistoryFileExist(t *testing.T) {
	f, err := os.Open(HistoryFilepath)
	defer f.Close()
	if nil == err {
		devnetDB.ReconstructFromFile(f)
		log.Info("new DB:", devnetDB)

	} else {
		log.Warn(HistoryFilepath, ":", err)
	}

}
*/

func TestOpenHistoryFileEmpty(t *testing.T) {
	devnetDB = make([]PingResult, 0)
	os.Remove(HistoryFilepath)
	f, err := os.Open(HistoryFilepath)
	if nil == err {
		devnetDB.ReconstructFromFile(f)
		defer f.Close()
	}
	log.Warn(HistoryFilepath, ":", err)
}
