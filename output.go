package main

import (
	"fmt"
	"strings"
	"time"
)

// ReportPingResultJSON is a struct convert from PingResult to desire json output struct
type DataPoint1MinResultJSON struct {
	Submitted  int    `json:"submitted"`
	Confirmed  int    `json:"confirmed"`
	Loss       string `json:"loss"`
	Mean       int    `json:"mean_ms"`
	TimeStamp  string `json:"ts"`
	ErrorCount int    `json:"error_count"`
	Error      string `json:"error"`
}

// SlackText slack structure
type SlackText struct {
	SText string `json:"text"`
	SType string `json:"type"`
}

// Block slack structure
type Block struct {
	BlockType string    `json:"type"`
	BlockText SlackText `json:"text"`
}

// SlackPayload slack structure
type SlackPayload struct {
	Blocks []Block `json:"blocks"`
}

// DiscordPayload slack structure
type DiscordPayload struct {
	BotName      string `json:"username"`
	BotAvatarURL string `json:"avatar_url"`
	Content      string `json:"content"`
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

func To1MinWindowJson(r *PingResult) DataPoint1MinResultJSON {
	// Check result
	jsonResult := DataPoint1MinResultJSON{Submitted: r.Submitted, Confirmed: r.Confirmed, Mean: int(r.Mean), Error: ErrorsToString(r.Error)}
	loss := fmt.Sprintf("%3.1f%s", r.Loss, "%")
	jsonResult.Loss = loss
	ts := time.Unix(r.TimeStamp, 0)
	jsonResult.TimeStamp = ts.Format(time.RFC3339)
	return jsonResult
}

// PingResultToJson convert PingSatistic to  RingResult Json format for API
func PingResultToJson(stat *PingSatistic) DataPoint1MinResultJSON {
	_, mean, _, _, _ := stat.TimeMeasure.Statistic()
	var errorShow string
	loss := fmt.Sprintf("%3.1f%s", stat.Loss*100, "%")
	if stat.Count == 0 {
		errorShow = "No Data"
		loss = fmt.Sprintf("%3.1f%s", float64(0), "%")
	}

	jsonResult := DataPoint1MinResultJSON{Submitted: int(stat.Submitted), Confirmed: int(stat.Confirmed), Mean: int(mean), ErrorCount: int(len(stat.Errors)), Error: errorShow}

	jsonResult.Loss = loss
	ts := time.Unix(stat.TimeStamp, 0).UTC()
	jsonResult.TimeStamp = ts.Format(time.RFC3339)
	return jsonResult
}

// ReportPayload get the report within specified minutes
func (s *SlackPayload) ReportPayload(c Cluster, data *GroupsAllStatistic, globalSatistic GlobalStatistic, hideKeywords []string, messageMemo string) {
	// Header Block
	headerText := fmt.Sprintf("total-submitted: %3.0f, total-confirmed:%3.0f, average-loss:%3.1f%s\n memo:%s",
		globalSatistic.Submitted,
		globalSatistic.Confirmed,
		globalSatistic.Loss*100, "%",
		messageMemo)
	header := Block{
		BlockType: "section",
		BlockText: SlackText{
			SType: "mrkdwn",
			SText: headerText,
		},
	}
	s.Blocks = append(s.Blocks, header)
	// BodyBlock
	body := Block{}
	records := reportRecordBlock(data)
	description := "( Submitted, Confirmed, Loss, min/mean/max/stddev ms )"
	//memo := "* rpc error : context deadline exceeded does not count as a transaction\n** BlockhashNotFound error does not show in Error List"
	memo := "*BlockhashNotFound do not count as a transaction\n"
	errorRecords := reportErrorBlock(data, hideKeywords)
	body = Block{
		BlockType: "section",
		BlockText: SlackText{SType: "mrkdwn", SText: fmt.Sprintf("```%s\n%s\n%s\n%s```", description, records, memo, errorRecords)},
	}
	s.Blocks = append(s.Blocks, body)
}

func (s *SlackPayload) AlertPayload(conf ClusterConfig, gStat *GlobalStatistic, errorStistic map[string]int, thresholdAdj float64, hideKeywords []string, messageMemo string) {
	var text, timeStatis string
	if gStat.TimeStatistic.Stddev <= 0 {
		timeStatis = fmt.Sprintf(" %d/%3.0f/%d/%s ", gStat.TimeStatistic.Min, gStat.TimeStatistic.Mean, gStat.TimeStatistic.Max, "NaN")
	} else {
		timeStatis = fmt.Sprintf(" %d/%3.0f/%d/%3.0f ", gStat.TimeStatistic.Min, gStat.TimeStatistic.Mean, gStat.TimeStatistic.Max, gStat.TimeStatistic.Stddev)
	}
	errsorStatis := ""
	for k, v := range errorStistic {
		if !PingResultError(k).IsInErrorList(AlertErrorExceptionList) {
			errsorStatis = fmt.Sprintf("%s%s(%d)", errsorStatis, PingResultError(k).Short(), v)
		}
	}
	for _, w := range hideKeywords {
		errsorStatis = strings.ReplaceAll(errsorStatis, w, "")
	}

	text = fmt.Sprintf("{ hostname: %s, memo: %s ,submitted: %3.0f, confirmed:%3.0f, loss: %3.1f%s, confirmation: min/mean/max/stddev = %s, next_threshold:%3.0f%s, error: %s}",
		conf.HostName, messageMemo, gStat.Submitted, gStat.Confirmed, gStat.Loss*100, "%", timeStatis, thresholdAdj, "%", errsorStatis)

	header := Block{
		BlockType: "section",
		BlockText: SlackText{
			SType: "mrkdwn",
			SText: text,
		},
	}
	s.Blocks = append(s.Blocks, header)
}

func (s *SlackPayload) FailoverAlertPayload(conf ClusterConfig, endpoint FailoverEndpoint, workerNum int) {
	text := fmt.Sprintf("{ hostname: %s, cluster:%s, worker:%d, msg:%s}",
		conf.HostName, conf.Cluster, workerNum,
		fmt.Sprintf("failover to %s", endpoint.Endpoint))

	header := Block{
		BlockType: "section",
		BlockText: SlackText{
			SType: "mrkdwn",
			SText: text,
		},
	}
	s.Blocks = append(s.Blocks, header)
}

func reportRecordBlock(data *GroupsAllStatistic) string {
	text := ""
	timeStatis := ""
	for _, ps := range data.PingStatisticList {
		if ps.TimeStatistic.Stddev <= 0 {
			timeStatis = fmt.Sprintf(" %d/%3.0f/%d/%s ", ps.TimeStatistic.Min, ps.TimeStatistic.Mean, ps.TimeStatistic.Max, "NaN")
		} else {
			timeStatis = fmt.Sprintf(" %d/%3.0f/%d/%3.0f ", ps.TimeStatistic.Min, ps.TimeStatistic.Mean, ps.TimeStatistic.Max, ps.TimeStatistic.Stddev)
		}
		lossPercentage := ps.Loss * 100
		if ps.Count > 0 {
			text = fmt.Sprintf("%s( %3.0f, %3.0f, %3.1f%s, %s )\n", text, ps.Submitted, ps.Confirmed, lossPercentage, "%", timeStatis)
		}
	}
	return text
}

// ReportPayload get the report within specified minutes
func (s *DiscordPayload) ReportPayload(c Cluster, data *GroupsAllStatistic, globalSatistic GlobalStatistic, hideKeywords []string, messageMemo string) {
	summary := fmt.Sprintf("**total-submitted: %3.0f  total-confirmed: %3.0f average-loss: %3.1f%s**\nmemo: %s",
		globalSatistic.Submitted,
		globalSatistic.Confirmed,
		globalSatistic.Loss*100, "%",
		messageMemo)
	header := "( Submitted, Confirmed, Loss, min/mean/max/stddev ms )"
	records := reportRecordBlock(data)
	memo := "*BlockhashNotFound do not count as a transaction\n"
	errorRecords := reportErrorBlock(data, hideKeywords)
	s.Content = fmt.Sprintf("%s\n```%s\n%s\n%s\n%s```", summary, header, records, memo, errorRecords)
}

// AlertPayload get the report within specified minutes
func (s *DiscordPayload) AlertPayload(conf ClusterConfig, gStat *GlobalStatistic, errorStistic map[string]int, thresholdAdj float64, hideKeywords []string, messageMemo string) {
	var timeStatis string
	if gStat.TimeStatistic.Stddev <= 0 {
		timeStatis = fmt.Sprintf(" %d/%3.0f/%d/%s ", gStat.TimeStatistic.Min, gStat.TimeStatistic.Mean, gStat.TimeStatistic.Max, "NaN")
	} else {
		timeStatis = fmt.Sprintf(" %d/%3.0f/%d/%3.0f ", gStat.TimeStatistic.Min, gStat.TimeStatistic.Mean, gStat.TimeStatistic.Max, gStat.TimeStatistic.Stddev)
	}
	errsorStatis := ""
	unconfirmedTx := 0
	for k, v := range errorStistic {
		if PingResultError(k).IsInErrorList(AlertErrorExceptionList) {
			continue
		}

		if UnconfirmedTx.IsIdentical(PingResultError(k)) {
			unconfirmedTx++
		} else {
			errsorStatis = fmt.Sprintf("%s%s(%d)", errsorStatis, PingResultError(k).Short(), v)
		}
	}
	errsorStatis += fmt.Sprintf("\n*(count:%d) Txs couldn't be confirmed\n", unconfirmedTx)
	for _, w := range hideKeywords {
		errsorStatis = strings.ReplaceAll(errsorStatis, w, "")
	}

	text := fmt.Sprintf("```{ hostname: %s, memo: %s, submitted: %3.0f, confirmed:%3.0f, loss: %3.1f%s, confirmation: min/mean/max/stddev = %s, next_threshold:%3.0f%s, error: %s}```",
		conf.HostName, messageMemo, gStat.Submitted, gStat.Confirmed, gStat.Loss*100, "%", timeStatis, thresholdAdj, "%", errsorStatis)
	s.Content = text
}

// FailoverAlertPayload get the report within specified minutes
func (s *DiscordPayload) FailoverAlertPayload(conf ClusterConfig, endpoint FailoverEndpoint, workerNum int) {
	s.Content = fmt.Sprintf("```{ hostname: %s, cluster:%s, worker:%d, msg:%s}```",
		conf.HostName, conf.Cluster, workerNum,
		fmt.Sprintf("failover to %s", endpoint.Endpoint))

}
func reportErrorBlock(data *GroupsAllStatistic, hideKeywords []string) string {
	var exceededText, errorText, blackHashText, unconfirmedTx string
	if len(data.GlobalErrorStatistic) == 0 {
		return ""
	}
	for k, v := range data.GlobalErrorStatistic {

		if RPCServerDeadlineExceeded.IsIdentical(PingResultError(k)) {
			exceededText = fmt.Sprintf("*(count:%d) RPC Server context deadline exceed\n", v)
		} else if BlockhashNotFound.IsIdentical(PingResultError(k)) {
			blackHashText = fmt.Sprintf("*(count:%d) BlockhashNotFound\n", v)
		} else if UnconfirmedTx.IsIdentical(PingResultError(k)) {
			unconfirmedTx = fmt.Sprintf("*(count:%d) Txs couldn't be confirmed\n", v)
		} else {
			errorText = fmt.Sprintf("%s\n(count: %d) %s\n", errorText, v, k)
		}
	}
	for _, w := range hideKeywords {
		errorText = strings.ReplaceAll(errorText, w, "")
	}
	if len(data.GlobalErrorStatistic) > 0 {
		return fmt.Sprintf("Error List:\n%s%s%s%s", exceededText, blackHashText, unconfirmedTx, errorText)
	}
	return ""
}

func reportRawErrorBlock(data *GroupsAllStatistic) string {
	var errorText string
	if len(data.GlobalErrorStatistic) == 0 {
		return ""
	}
	for k, v := range data.GlobalErrorStatistic {
		errorText = fmt.Sprintf("%s\n(count: %d) %s", errorText, v, k)
	}
	return fmt.Sprintf("Error List:%s", errorText)
}
