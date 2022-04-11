package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/parnurzeal/gorequest"
)

type SlackTriggerEvaluation struct {
	LastLoss        float64
	CurrentLoss     float64
	ThresHoldIndex  int
	ThresHoldLevels []float64
	TrendAsc        bool
	FilePath        string
}

func NewSlackTriggerEvaluation() SlackTriggerEvaluation {
	s := SlackTriggerEvaluation{}
	s.FilePath = config.SlackReport.SlackAlert.LevelFilePath
	s.CurrentLoss = 0
	s.LastLoss = 0
	s.ThresHoldLevels = []float64{float64(config.SlackReport.SlackAlert.LossThreshold), float64(50), float64(75), float64(100)}
	s.FilePath = config.SlackReport.SlackAlert.LevelFilePath
	s.ThresHoldIndex = s.ReadFromFile()

	return s
}

func (s *SlackTriggerEvaluation) Update(currentLoss float64) {
	s.LastLoss = s.CurrentLoss
	s.CurrentLoss = currentLoss * 100
	if s.CurrentLoss > s.LastLoss {
		s.TrendAsc = true
	} else {
		s.TrendAsc = false
	}
}

func (s *SlackTriggerEvaluation) UpperLevel(loss float64) int {
	if loss <= s.ThresHoldLevels[0] {
		return 0
	}
	if loss >= s.ThresHoldLevels[len(s.ThresHoldLevels)-1] {
		return len(s.ThresHoldLevels) - 1
	}
	// > level 0
	for i, v := range s.ThresHoldLevels {
		if loss < v {
			return i
		}
	}
	return 0
}

func (s *SlackTriggerEvaluation) WriteLevelToFile(level int) {
	if s.FilePath != "" {
		os.WriteFile(s.FilePath, []byte(strconv.Itoa(level)), 0777)
		log.Println("WriteLevelToFile : ", strconv.Itoa(level), " to ", s.FilePath)
	}
}

func (s *SlackTriggerEvaluation) ReadFromFile() int {
	if s.FilePath != "" {
		v, err := os.ReadFile(s.FilePath)
		if err != nil {
			return 0
		}
		level, err := strconv.Atoi(string(v))
		if err != nil {
			return 0
		}
		log.Println("ReadFromFile : ", level, " from ", s.FilePath)
		return level
	}
	return 0
}

// Doing rule here
func (s *SlackTriggerEvaluation) ShouldSend() bool {
	// less than initial threadhold, set Index to Initial level
	if s.CurrentLoss < s.ThresHoldLevels[0] {
		s.ThresHoldIndex = 0
		s.WriteLevelToFile(0)
		log.Println("ThreadHoldLevels Down To :", s.ThresHoldIndex, " ShouldSend", false)
		return false
	}
	// NextLevel UP
	if s.CurrentLoss > s.ThresHoldLevels[s.ThresHoldIndex] {
		s.ThresHoldIndex = s.UpperLevel(s.CurrentLoss)
		s.WriteLevelToFile(s.ThresHoldIndex)
		log.Println("ThreadHoldLevels Up To :", s.ThresHoldIndex, " ShouldSend", true)
		return true
	}
	// NextLevel Down
	if s.ThresHoldIndex >= 2 {
		if s.CurrentLoss < s.ThresHoldLevels[s.ThresHoldIndex-2] {
			if s.ThresHoldIndex-2 > 0 { // Big decrease
				s.ThresHoldIndex = s.ThresHoldIndex - 1
				s.WriteLevelToFile(s.ThresHoldIndex)
				log.Println("ThreadHoldLevels Down To :", s.ThresHoldIndex, " ShouldSend", true)
				return true
			} else { // go back to normal level
				s.ThresHoldIndex = 0
				s.WriteLevelToFile(0)
				log.Println("ThreadHoldLevels Down To :", s.ThresHoldIndex, " ShouldSend", false)
				return false
			}
		}
	}
	return false
}

func slackSend(cluster Cluster, globalStat *GlobalStatistic, globalErrorStatistic map[string]int, threadhold float64) {
	loss := globalStat.Loss * 100
	if loss > float64(config.SlackReport.SlackAlert.LossThreshold) {
		payload := SlackPayload{}
		payload.AlertPayload(cluster, globalStat, globalErrorStatistic, threadhold)
		err := SlackSend(config.SlackReport.SlackAlert.WebHook, &payload)
		if err != nil {
			log.Println("SlackSend Error:", err)
		}
	}
}

//SlackSend send the payload to then Slack Channel
func SlackSend(webhookURL string, payload *SlackPayload) []error {
	data, err := json.Marshal(*payload)
	if err != nil {
		return []error{fmt.Errorf("marshal payload error. Status:%s", err.Error())}
	}
	request := gorequest.New()
	resp, _, errs := request.Post(webhookURL).Send(string(data)).End()

	if errs != nil {
		return errs
	}
	if resp.StatusCode >= 400 {
		log.Println(resp.StatusCode, " payload:", string(data))
		return []error{fmt.Errorf("slack sending msg. Status: %v", resp.Status)}
	}

	log.Println("Slack send successfully! =>", webhookURL)
	return nil
}
