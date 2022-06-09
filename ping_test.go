package main

import (
	"fmt"
	"testing"
	"time"
)

var sch1 = PingResult{
	TimeStamp: time.Now().UTC().Unix(),
	Cluster:   "Devnet",
	Hostname:  "solana-ping-api",
	PingType:  "report",
	Submitted: 10,
	Confirmed: 9,
	Max:       12000,
	Mean:      8000,
	Min:       500,
	Stddev:    100,
	TakeTime:  90,
	Error:     []string{},
}

var hook = "https://hooks.slack.com/services/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"

func TestIsIdentical(t *testing.T) {
	su503 := PingResultError(ServiceUnavilable503Text)
	blackhash := PingResultError(BlockhashNotFoundText)
	if !ServiceUnavilable503.IsIdentical(su503) {
		t.Fatal("ServiceUnavilable503 should be true")
	}
	if BlockhashNotFound.IsIdentical(su503) {
		t.Fatal("BlockhashNotFound should be false")
	}
	if !blackhash.IsInErrorList(StatisticErrorExceptionList) {
		t.Fatal("blackhash should be in the list")
	}
	if su503.IsInErrorList(StatisticErrorExceptionList) {
		t.Fatal("503 should not be in the list")
	}
	if !(su503.Short() == ServiceUnavilable503.Short) {
		t.Fatal("503 short is not correct")
	} else {
		fmt.Println(su503.Short())
	}
}

// func TestParse(t *testing.T) {
// 	pings := []PingResult{sch1}
// 	avg := generateStatisticData(pings)
// 	payload := SlackPayload{}
// 	payload.ToReportPayload(MainnetBeta, pings, avg)
// 	err := SlackSend(hook, &payload)
// 	if err != nil {
// 		log.Print(err)
// 	}

// }

/*
func TestConfig(t *testing.T) {
	config := loadConfig()
	log.Println(config.Clusters)
	log.Println(config.ServerIP)
	log.Println(config.SolanaConfig)
	log.Println(config.SolanaPing)
	log.Println(config.Slack)
	log.Println(config.Cleaner)
}
*/

/*
func TestSlack(t *testing.T) {

	payload := SlackPayload{}

	payload.GetReportPayload(Devnet)

	errs := SlackSend(config.Slack.WebHook, &payload)
	if errs != nil {
		t.Error(errs)
	}

}
*/
var output = `Source Account: 8Juu8gXPHCKougXvkg3ZY8HdJcBdqFch7hHKjKFNWTFV

msg BD5heio2zqXdNuN1i7hia6dKbLVq7RvYt9HanYdj8nWD
✅ 1 lamport(s) transferred: seq=0   time= 773ms signature=2twfwLjCj7eLxXPDkyfPqJFP66Jb3Npkb6nTVeJXXZC5rThYp9Mn67SX8Pk8km5SAyQwYdasrK7cUP4JJN2BdXYr
msg BD5heio2zqXdNuN1i7hia6dKbLVq7RvYt9HanYdj8nWD
✅ 1 lamport(s) transferred: seq=1   time=1106ms signature=Nbyz2bzmb2FNqdmJoP18KoHeNE8ByvCHh2D4tiDsWxn7Z5BoWXsA5iGrhgjbqRDe5FQy5gafHfcaXw73c4spYod
msg BD5heio2zqXdNuN1i7hia6dKbLVq7RvYt9HanYdj8nWD
✅ 1 lamport(s) transferred: seq=2   time= 773ms signature=2JXmQNaLW2FM3w9uMmTJqRZiHJkrZay1e6B7ti9mAWebgDWqCek6YoHmrLoLu7wNqSLuHSmgjpKv5ERoq4HYyMMz

--- transaction statistics ---
3 transactions submitted, 3 transactions confirmed, 0.0% transaction loss
confirmation min/mean/max/stddev = 773/884/1106/192 ms
 `
