package main

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type PingResult struct {
	Cluster
	Hostname            string
	Submitted           uint64
	Confirmed           uint64
	Loss                float64
	ConfirmationMessage string
	TimeStamp           int64
	ErrorMessage        error
}

type PingResultJson struct {
	Hostname            string `json:"hostname"`
	Cluster             `json:"cluster"`
	Submitted           uint64 `json:"submitted"`
	Confirmed           uint64 `json:"confirmed"`
	Loss                string `json:"loss"`
	ConfirmationMessage string `json:"confirmation"`
	TimeStamp           string `json:"ts"`
	ErrorMessage        string `json:"error"`
}

const (
	RegexpSubmitted    = "[0-9]+\\stransactions submitted"
	RegexpConfirmed    = "[0-9]+\\stransactions confirmed"
	RegexpLoss         = "([0-9]*[.])?[0-9]%\\stransaction loss"
	RegexpConfirmation = "min/mean/max/stddev\\s*=\\s*[\\s\\S]*ms"
)

func findingPattern(reg *regexp.Regexp, output string) (string, error) {
	loc := reg.FindIndex([]byte(output))
	if nil == loc {
		return "", FindIndexNotFound
	}
	return output[loc[0]:loc[1]], nil
}

func (r *PingResult) parsePingOutput(output string) error {

	// Submitted
	reg := regexp.MustCompile(RegexpSubmitted)
	subSentence, err := findingPattern(reg, output)
	if err != nil {
		r.TimeStamp = time.Now().UTC().Unix()
		r.ErrorMessage = err
		return err
	}
	tmp := strings.Split(subSentence, " ")
	n, err := strconv.ParseUint(tmp[0], 10, 64)
	if err != nil {
		log.Error("parse transactions confirmed error ", subSentence)
		r.ErrorMessage = err
		return errors.New("Parse Output Error")
	}
	r.Submitted = n

	// Confirmed
	reg = regexp.MustCompile(RegexpConfirmed)
	subSentence, err = findingPattern(reg, output)
	if err != nil {
		r.TimeStamp = time.Now().Unix()
		r.ErrorMessage = err
		return err
	}
	tmp = strings.Split(subSentence, " ")
	n, err = strconv.ParseUint(tmp[0], 10, 64)
	if err != nil {
		log.Error("parse transactions confirmed error ", subSentence)
		r.ErrorMessage = err
		return ConvertWrongType
	}
	r.Confirmed = n

	// loss
	reg = regexp.MustCompile(RegexpLoss)
	subSentence, err = findingPattern(reg, output)
	if err != nil {
		r.TimeStamp = time.Now().Unix()
		r.ErrorMessage = err
		return err
	}
	tmp = strings.Split(subSentence, "%")
	if len(tmp) != 2 {
		r.ErrorMessage = ParseSplitError
		return ParseSplitError
	}
	lossval, err := strconv.ParseFloat(tmp[0], 64)
	if err != nil {
		log.Error("parse transactions loss error ", subSentence)
		r.ErrorMessage = ConvertWrongType
		return ConvertWrongType
	}
	r.Loss = lossval

	// Confirmation
	reg = regexp.MustCompile(RegexpConfirmation)
	subSentence, err = findingPattern(reg, output)
	if err != nil {
		r.TimeStamp = time.Now().Unix()
		r.ErrorMessage = err
		return err
	}
	if len(subSentence) <= 0 {
		r.TimeStamp = time.Now().Unix()
		r.ErrorMessage = ParseSplitError
		return ParseSplitError
	}
	r.TimeStamp = time.Now().Unix()
	r.ConfirmationMessage = subSentence
	r.ErrorMessage = nil
	return nil
}

func (r *PingResult) ConvertToJoson() (PingResultJson, error) {
	// Check result
	jsonResult := PingResultJson{Hostname: r.Hostname, Cluster: r.Cluster, Submitted: r.Submitted, Confirmed: r.Confirmed,
		ConfirmationMessage: r.ConfirmationMessage}
	if nil == r.ErrorMessage {
		jsonResult.ErrorMessage = ""
	}
	if r.Submitted == 0 && r.Confirmed == 0 && len(r.ConfirmationMessage) == 0 {
		return jsonResult, ResultInvalid

	}

	loss := fmt.Sprintf("%3.1f%s", r.Loss, "%")

	jsonResult.Loss = loss
	ts := time.Unix(r.TimeStamp, 0)
	jsonResult.TimeStamp = ts.Format(time.RFC3339)
	return jsonResult, nil
}

// Memo: Below regex is not working for e2
//[0-9]+(?=\stransactions submitted)
//[0-9]+(?=\stransactions confirmed)
//[0-9]+[.]*[0-9]*%(?= transaction loss)
//confirmation[\s\S]*ms
