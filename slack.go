package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/parnurzeal/gorequest"
)

const Header = "( Submitted, Confirmed, Loss, min/mean/max/stddev ms )"

type SlackText struct {
	SText string `json:"text"`
	SType string `json:"type"`
}

type Block struct {
	BlockType string    `json:"type"`
	BlockText SlackText `json:"text"`
}

type SlackPayload struct {
	Blocks []Block `json:"blocks"`
}

type averageResult struct {
	Loss  float64
	Count int
}

var lastReporUnixTime int64
var ClusterToReport = []Cluster{Testnet, Devnet, MainnetBeta}

func resultMarkdown(pr []PingResult) (string, averageResult, error) {
	if len(pr) <= 0 {
		return "()", averageResult{}, NoPingResultRecord
	}

	text := ""
	var sumLoss float64 = 0
	count := 0
	for _, e := range pr {
		cmsg := strings.Split(e.ConfirmationMessage, " ")
		var confdata string
		if len(cmsg) < 4 {
			log.Error("split confirmationMessage error:", cmsg)
			confdata = e.ConfirmationMessage
		} else {
			confdata = cmsg[2]
		}

		text = fmt.Sprintf("%s( %d, %d, %3.1f, %s , %d )\n", text, e.Submitted, e.Confirmed, e.Loss, confdata, e.TimeStamp)
		sumLoss += e.Loss
		count++
	}
	avgLoss := sumLoss / float64(count)
	return text, averageResult{Loss: avgLoss, Count: count}, nil
}

func (s *SlackPayload) GetReportPayload(c Cluster) {
	currentTime := time.Now().Unix()
	result := []PingResult{}
	switch c {
	case MainnetBeta:
		result = mainnetBetaDB.GetTimeAfter(lastReporUnixTime)
	case Testnet:
		result = testnetDB.GetTimeAfter(lastReporUnixTime)
	case Devnet:
		result = devnetDB.GetTimeAfter(lastReporUnixTime)
	default:
		result = devnetDB.GetTimeAfter(lastReporUnixTime)
	}

	records, avg, err := resultMarkdown(result)
	lastReporUnixTime = currentTime
	log.Info("GetReportPayload fetch ", len(result), " from ", c)

	body := Block{}
	if err == nil {
		header := Block{BlockType: "section", BlockText: SlackText{SType: "mrkdwn", SText: fmt.Sprintf("%d results availible %s\n", avg.Count, Header)}}
		s.Blocks = append(s.Blocks, header)

		body = Block{BlockType: "section", BlockText: SlackText{SType: "mrkdwn", SText: fmt.Sprintf("```%s```\n", records)}}
		s.Blocks = append(s.Blocks, body)

		footer := Block{BlockType: "section", BlockText: SlackText{SType: "mrkdwn", SText: fmt.Sprintf("%s\n", fmt.Sprintf("Average Loss %3.1f%s", avg.Loss, "%"))}}
		s.Blocks = append(s.Blocks, footer)
	}

}

func SlackSend(webhookUrl string, payload *SlackPayload) []error {
	data, err := json.Marshal(*payload)

	if err != nil {
		return []error{fmt.Errorf("marshal payload error. Status:%s", err.Error())}
	}

	request := gorequest.New()
	resp, _, errs := request.Post(webhookUrl).Send(string(data)).End()

	if errs != nil {
		log.Error(err)
		return errs
	}

	if resp.StatusCode >= 400 {
		return []error{fmt.Errorf("slack sending msg. Status: %v", resp.Status)}
	}

	log.Info("SlackSend->")

	return nil
}

func SlackReportService() {
	log.Info("-- Start SlackReportService --")
	for _, c := range ClusterToReport {
		go SlackReport(c)
	}
}

func SlackReport(c Cluster) {
	payload := SlackPayload{}
	payload.GetReportPayload(c)
	if len(payload.Blocks) <= 0 {
		time.Sleep(time.Duration(config.Slack.ReportTime) * time.Second)
	}
	SlackSend(config.Slack.WebHook, &payload)
}
