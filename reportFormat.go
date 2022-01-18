package main

import (
	"fmt"
	"log"
	"strings"
)

//SlackText slack structure
type SlackText struct {
	SText string `json:"text"`
	SType string `json:"type"`
}

//Block slack structure
type Block struct {
	BlockType string    `json:"type"`
	BlockText SlackText `json:"text"`
}

//SlackPayload slack structure
type SlackPayload struct {
	Blocks []Block `json:"blocks"`
}

//statisticResult for caucualate report
type statisticResult struct {
	NumOfRecords int
	Submiited    float64
	Confirmed    float64
	Loss         float64
	Count        int
	ErrCount     int
}

//ToPayload get the report within specified minutes
func (s *SlackPayload) ToPayload(c Cluster, data []PingResult, stats statisticResult) {
	records, err := reportBody(data, stats)
	description := "( Submitted, Confirmed, Loss, min/mean/max/stddev ms )"
	body := Block{}
	if err == nil {
		header := Block{
			BlockType: "section",
			BlockText: SlackText{
				SType: "mrkdwn",
				SText: fmt.Sprintf("%d results availible in %s. %s\n",
					stats.Count, c, fmt.Sprintf("Average Loss %3.1f%s", stats.Loss, "%")),
			},
		}
		s.Blocks = append(s.Blocks, header)

		body = Block{
			BlockType: "section",
			BlockText: SlackText{SType: "mrkdwn", SText: fmt.Sprintf("```%s\n%s```", description, records)},
		}
		s.Blocks = append(s.Blocks, body)
	}
}

func generateData(pr []PingResult) statisticResult {
	var sumSub, sumConf, sumLoss float64
	count := 0
	errCount := 0
	for _, e := range pr {
		sumSub += float64(e.Submitted)
		sumConf += float64(e.Confirmed)
		sumLoss += e.Loss
		count++
	}
	avgLoss := sumLoss / float64(count)
	avgSub := sumSub / float64(count)
	avgConf := sumConf / float64(count)
	return statisticResult{NumOfRecords: len(pr), Submiited: avgSub, Confirmed: avgConf, Loss: avgLoss, Count: count, ErrCount: errCount}
}

func reportBody(pr []PingResult, st statisticResult) (string, error) {
	if st.Count <= 0 {
		return "()", NoPingResultRecord
	}

	text := ""
	for _, e := range pr {
		cmsg := strings.Split(e.ConfirmationMessage, " ")
		var confdata string
		if len(cmsg) < 4 {
			log.Println("split confirmationMessage error:", cmsg, " PingResult=>", e)
			confdata = e.ConfirmationMessage
		} else {
			confdata = cmsg[2]
		}
		text = fmt.Sprintf("%s( %d, %d, %3.1f, %s )\n", text, e.Submitted, e.Confirmed, e.Loss, confdata)
	}

	return text, nil
}
