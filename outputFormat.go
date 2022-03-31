package main

import (
	"fmt"
	"log"
	"strings"
	"time"
)

//ReportResultJSON is a struct convert from PingResult to desire json output struct
type ReportResultJSON struct {
	Hostname            string `json:"hostname"`
	Submitted           int    `json:"submitted"`
	Confirmed           int    `json:"confirmed"`
	Loss                string `json:"loss"`
	ConfirmationMessage string `json:"confirmation"`
	TimeStamp           string `json:"ts"`
	Error               string `json:"error"`
}

//DataPoint1MinJson is a struct return to api
type DataPoint1MinJson struct {
	NumOfDataPoint int                       `json:"num_data_point"`
	NumOfNoData    int                       `json:"num_of_nodata"`
	Data           []DataPoint1MinResultJSON `json:"data"`
}

//ReportPingResultJSON is a struct convert from PingResult to desire json output struct
type DataPoint1MinResultJSON struct {
	Submitted int    `json:"submitted"`
	Confirmed int    `json:"confirmed"`
	Loss      string `json:"loss"`
	Mean      int    `json:"mean_ms"`
	TimeStamp string `json:"ts"`
	Error     string `json:"error"`
}

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
	Submitted float64
	Confirmed float64
	Loss      float64
	Count     int
	ErrList   map[string]int
	ErrCount  int
}

func confirmationMessage(pr PingResult) string {
	if pr.Stddev <= 0 {
		return fmt.Sprintf(" %d/%d/%d/%s ", int(pr.Min), int(pr.Mean), int(pr.Max), "NaN")
	}
	return fmt.Sprintf(" %d/%d/%d/%d ", int(pr.Min), int(pr.Mean), int(pr.Max), int(pr.Stddev))
}

func ErrorsToString(errs []string) (errsString string) {
	for i, e := range errs {
		if i == 0 {
			errsString = e
		} else {
			errsString = errsString + ";" + e
		}
	}
	return
}

//ToReportJoson convert PingResult to Json Format
func ToReportJoson(r *PingResult) ReportResultJSON {
	// Check result
	jsonResult := ReportResultJSON{Hostname: r.Hostname, Submitted: r.Submitted, Confirmed: r.Confirmed,
		ConfirmationMessage: confirmationMessage(*r), Error: ErrorsToString(r.Error)}
	loss := fmt.Sprintf("%3.1f%s", r.Loss, "%")
	jsonResult.Loss = loss
	ts := time.Unix(r.TimeStamp, 0)
	jsonResult.TimeStamp = ts.Format(time.RFC3339)
	return jsonResult
}

func To1MinWindowJson(r *PingResult) DataPoint1MinResultJSON {
	// Check result
	jsonResult := DataPoint1MinResultJSON{Submitted: r.Submitted, Confirmed: r.Confirmed, Mean: int(r.Mean), Error: ErrorsToString(r.Error)}
	loss := fmt.Sprintf("%3.1f%s", r.Loss, "%")
	jsonResult.Loss = loss
	ts := time.Unix(r.TimeStamp, 0)
	jsonResult.TimeStamp = ts.Format(time.RFC3339)
	return jsonResult
}

//ToReportPayload get the report within specified minutes
func (s *SlackPayload) ToReportPayload(c Cluster, data []PingResult, stats *statisticResult) {
	records, err := reportBody(data, stats)
	if err == NoPingResultRecord {
		records = "no data availible"
	}
	description := "( Submitted, Confirmed, Loss, min/mean/max/stddev ms )"
	body := Block{}
	if err == nil {
		header := Block{
			BlockType: "section",
			BlockText: SlackText{
				SType: "mrkdwn",
				SText: fmt.Sprintf("%d results availible in %s. %s\n",
					stats.Count, c, fmt.Sprintf("Average Loss %3.1f%s", stats.Loss*100, "%")),
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

// ToAlertPayload get the report within specified minutes
func (s *SlackPayload) ToAlertPayload(c Cluster, data []PingResult, stats *statisticResult) {
	records, err := alertBody(stats)
	if err == NoPingResultRecord {
		records = "no data availible"
	}
	body := Block{}
	lossPercent := float64(100)
	if stats.Submitted > 0 {
		lossPercent = (float64(stats.Submitted-stats.Confirmed) / float64(stats.Submitted)) * 100
	}
	slackText := fmt.Sprintf("Ping Alert! %s reports %3.1f%s loss.",
		config.HostName, lossPercent, "%")
	if err == nil {
		header := Block{
			BlockType: "section",
			BlockText: SlackText{
				SType: "mrkdwn",
				SText: slackText,
			},
		}
		s.Blocks = append(s.Blocks, header)
		body = Block{
			BlockType: "section",
			BlockText: SlackText{SType: "mrkdwn", SText: fmt.Sprintf("```%s```", records)},
		}
		s.Blocks = append(s.Blocks, body)
	}
}

func generateStatisticData(pr []PingResult) *statisticResult {
	var sumSub, sumConf float64
	count := 0
	result := statisticResult{}
	result.ErrList = make(map[string]int) // initiialize map
	for _, e := range pr {
		sumSub += float64(e.Submitted)
		sumConf += float64(e.Confirmed)
		count++
		if len(e.Error) > 0 {
			for _, e := range e.Error {
				result.ErrList[string(e)] = result.ErrList[string(e)] + 1
			}
		}
	}
	if count <= 0 { // return an default result
		return &result
	}

	result.Submitted = float64(sumSub)
	result.Confirmed = float64(sumConf)
	if result.Submitted > 0 {
		result.Loss = (result.Submitted - result.Confirmed) / result.Submitted
	} else {
		result.Loss = 100
	}

	for _, c := range result.ErrList {
		result.ErrCount = result.ErrCount + c
	}
	result.Count = count
	return &result
}

func reportBody(pr []PingResult, st *statisticResult) (string, error) {
	if st.Count <= 0 {
		return "()", NoPingResultRecord
	}
	text := ""
	for _, e := range pr {
		cmsg := confirmationMessage(e)
		loss := float64(100)
		if e.Submitted > 0 {
			loss = (float64(e.Submitted-e.Confirmed) / float64(e.Submitted)) * 100
		}
		failCount := check429fail(e)
		if failCount > 0 {
			text = fmt.Sprintf("%s( %d, %d, %3.1f, %s )(%d 429-Too-Many-Requests rejects)\n", text, e.Submitted, e.Confirmed, loss, cmsg, failCount)
		} else {
			text = fmt.Sprintf("%s( %d, %d, %3.1f, %s )\n", text, e.Submitted, e.Confirmed, loss, cmsg)
		}

		log.Println("reportBody:", text)
	}
	return text, nil
}

func alertBody(st *statisticResult) (string, error) {
	if st.Count <= 0 {
		return "()", NoPingResultRecord
	}
	text := ""

	loss := float64(100)
	if st.Submitted > 0 {
		loss = st.Loss * 100
	}
	errListDisplay := ""
	if len(st.ErrList) > 0 {
		for k, _ := range st.ErrList {
			if !strings.Contains(k, "-32002") {
				errListDisplay = errListDisplay + fmt.Sprintf("   *%s\n", k)
			}
		}
	}
	text = fmt.Sprintf("Submitted: %d, Confirmed: %d, Loss: %3.1f%s, Loss-Thredhold: %d%s, DataWindow: %d secs.\nError List:\n%s\n",
		int(st.Submitted), int(st.Confirmed), loss, "%", config.SlackAlert.LossThredhold, "%", config.SlackAlert.DataWindow, errListDisplay)
	log.Println("alertBody:", text)
	return text, nil
}

func generateDataPoint1Min(startTime int64, endTime int64, pr []PingResult) ([]DataPoint1MinResultJSON, int) {
	window := int64(60)
	datapoint1MinRet := []PingResult{}
	nodata := 0
	for periodend := endTime; periodend > startTime; periodend = periodend - window {
		count := 0
		sumOfMean := float64(0)
		countSuccess := 0
		windowResult := PingResult{}
		for _, result := range pr {
			if result.TimeStamp <= periodend && result.TimeStamp > periodend-window {
				windowResult.Submitted = windowResult.Submitted + result.Submitted
				windowResult.Confirmed = windowResult.Confirmed + result.Confirmed
				windowResult.TimeStamp = periodend // use the newest one for easier tracking
				windowResult.Hostname = result.Hostname
				if len(result.Error) <= 0 {
					sumOfMean = sumOfMean + float64(result.Mean)
					countSuccess++
				}

				count++
			}
		}
		if count == 0 {
			windowResult.Error = []string{"No Data"}
			windowResult.Submitted = 0
			windowResult.Confirmed = 0
			windowResult.Mean = 0
			windowResult.TimeStamp = periodend
			nodata++
		} else {
			if windowResult.Submitted > 0 {
				windowResult.Loss = (float64(windowResult.Submitted-windowResult.Confirmed) / float64(windowResult.Submitted)) * 100
			}
			if countSuccess > 0 {
				windowResult.Mean = int64(sumOfMean / float64(countSuccess))
			}
		}

		windowResult.PingType = string(DataPoint1Min)
		datapoint1MinRet = append(datapoint1MinRet, windowResult)
	}
	ret := []DataPoint1MinResultJSON{}
	for _, e := range datapoint1MinRet {
		ret = append(ret, To1MinWindowJson(&e))
	}
	return ret, nodata
}

func check429fail(pr PingResult) int {
	cnt := 0
	for _, er := range pr.Error {
		if strings.Contains(er, "get status code: 429") {
			cnt++
		}
	}
	return cnt

}
