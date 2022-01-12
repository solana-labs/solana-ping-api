package main

import (
	"errors"
	"testing"
)

var pingResult1 = PingResult{Hostname: "a", Submitted: 10, Confirmed: 10, Loss: 0, ConfirmationMessage: "nothing", ErrorMessage: nil}
var pingResult2 = PingResult{Hostname: "b", Submitted: 10, Confirmed: 0, Loss: 100, ConfirmationMessage: "nothing", ErrorMessage: nil}
var pingResult3 = PingResult{Hostname: "c", Submitted: 5, Confirmed: 5, Loss: 50, ConfirmationMessage: "nothing", ErrorMessage: errors.New("OooMyErr")}
var testpayload = `{"blocks":[{"type":"section","text":{"text":"10 results","type":"mrkdwn"}}]}`

func TestSlack(t *testing.T) {
	c := loadConfig()
	log.Info(c)
}

/*
func TestSlack(t *testing.T) {
	os.Remove(HistoryFilepathDevnet)
	devnetDB = make([]PingResult, 0)
	devnetDB.Add(pingResult1, Devnet)
	devnetDB.Add(pingResult2, Devnet)
	devnetDB.Add(pingResult3, Devnet)

	payload := SlackPayload{}

	payload.GetReportPayload(Devnet)

	errs := SlackSend(SolanaPingWebHook, &payload)
	if errs != nil {
		t.Error(errs)
	}

}

func TestSaveToFile(t *testing.T) {
	os.Remove(HistoryFilepath)
	devnetDB = make([]PingResult, 0)
	devnetDB.Add(pingResult1)
	devnetDB.Add(pingResult2)
	devnetDB.Add(pingResult3)
	f, err := os.OpenFile(HistoryFilepath, os.O_CREATE|os.O_RDWR, 0644)
	defer f.Close()
	if err != nil {
		t.FailNow()
	}
	err = devnetDB.SaveToFile(f)
	if err != nil {
		t.FailNow()
	}
}

func TestOpenHistoryFileExist(t *testing.T) {
	os.Remove(HistoryFilepath)
	devnetDB = make([]PingResult, 0)
	devnetDB.Add(pingResult1)
	devnetDB.Add(pingResult2)
	devnetDB.Add(pingResult3)
	SaveToFileHelper()
	f, err := os.Open(HistoryFilepath)
	defer f.Close()
	if nil == err {
		devnetDB.ReconstructFromFile(f)
		t.Log("new DB:", devnetDB)

	} else {
		t.Error(err)
	}
}

func TestOpenHistoryFileEmpty(t *testing.T) {
	devnetDB = make([]PingResult, 0)
	os.Remove(HistoryFilepath)
	f, err := os.Open(HistoryFilepath)
	defer f.Close()

	if nil == err {
		reconstructErr := devnetDB.ReconstructFromFile(f)
		if reconstructErr != nil {
			t.Error()
		}
	}
}
*/
