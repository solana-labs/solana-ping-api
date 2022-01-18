package main

import (
	"errors"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	regexpSubmitted    = "[0-9]+\\stransactions submitted"
	regexpConfirmed    = "[0-9]+\\stransactions confirmed"
	regexpLoss         = "([0-9]*[.])?[0-9]%\\stransaction loss"
	regexpConfirmation = "min/mean/max/stddev\\s*=\\s*[\\s\\S]*ms"
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
	reg := regexp.MustCompile(regexpSubmitted)
	subSentence, err := findingPattern(reg, output)
	if err != nil {
		r.TimeStamp = time.Now().UTC().Unix()
		r.Error = err.Error()
		return err
	}
	tmp := strings.Split(subSentence, " ")
	n, err := strconv.ParseUint(tmp[0], 10, 32)
	if err != nil {
		log.Println("parse transactions confirmed error ", subSentence)
		r.Error = err.Error()
		return errors.New("Parse Output Error")
	}
	r.Submitted = int(n)

	// Confirmed
	reg = regexp.MustCompile(regexpConfirmed)
	subSentence, err = findingPattern(reg, output)
	if err != nil {
		r.TimeStamp = time.Now().UTC().Unix()
		r.Error = err.Error()
		return err
	}
	tmp = strings.Split(subSentence, " ")
	n, err = strconv.ParseUint(tmp[0], 10, 32)
	if err != nil {
		log.Println("parse transactions confirmed error ", subSentence)
		r.Error = err.Error()
		return ConvertWrongType
	}
	r.Confirmed = int(n)

	// loss
	reg = regexp.MustCompile(regexpLoss)
	subSentence, err = findingPattern(reg, output)
	if err != nil {
		r.TimeStamp = time.Now().UTC().Unix()
		r.Error = err.Error()
		return err
	}
	tmp = strings.Split(subSentence, "%")
	if len(tmp) != 2 {
		r.Error = ParseSplitError.Error()
		return ParseSplitError
	}
	lossval, err := strconv.ParseFloat(tmp[0], 64)
	if err != nil {
		log.Println("parse transactions loss error ", subSentence)
		r.Error = ConvertWrongType.Error()
		return ConvertWrongType
	}
	r.Loss = lossval

	// Confirmation
	reg = regexp.MustCompile(regexpConfirmation)
	subSentence, err = findingPattern(reg, output)
	if err != nil {
		r.TimeStamp = time.Now().UTC().Unix()
		r.Error = err.Error()
		return err
	}
	if len(subSentence) <= 0 {
		r.TimeStamp = time.Now().UTC().Unix()
		r.Error = ParseSplitError.Error()
		return ParseSplitError
	}
	r.TimeStamp = time.Now().UTC().Unix()
	r.ConfirmationMessage = subSentence
	r.Error = ""
	return nil
}

// Memo: Below regex is not working for e2
//[0-9]+(?=\stransactions submitted)
//[0-9]+(?=\stransactions confirmed)
//[0-9]+[.]*[0-9]*%(?= transaction loss)
//confirmation[\s\S]*ms
