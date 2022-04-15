package main

import (
	"log"
	"os"
	"strconv"
)

type AlertTriggerEvaluation struct {
	LastLoss        float64
	CurrentLoss     float64
	ThresholdIndex  int
	ThresholdLevels []float64
	ThresholdAsc    bool
	FilePath        string
}

func NewAlertTriggerEvaluation() AlertTriggerEvaluation {
	s := AlertTriggerEvaluation{}
	s.FilePath = config.SlackReport.SlackAlert.LevelFilePath
	s.CurrentLoss = 0
	s.LastLoss = 0
	s.ThresholdLevels = []float64{float64(config.SlackReport.SlackAlert.LossThreshold), float64(50), float64(75), float64(100)}
	s.FilePath = config.SlackReport.SlackAlert.LevelFilePath
	s.ThresholdIndex = s.ReadFromFile()

	return s
}

func (s *AlertTriggerEvaluation) Update(currentLoss float64) {
	s.LastLoss = s.CurrentLoss
	s.CurrentLoss = currentLoss * 100
}

// get a threshold index which is 1 level higher than loss value
func (s *AlertTriggerEvaluation) UpLevel(loss float64) int {
	if loss < s.ThresholdLevels[0] {
		return 0
	}
	if loss >= s.ThresholdLevels[len(s.ThresholdLevels)-1] {
		return len(s.ThresholdLevels) - 1
	}
	// > level 0
	for i, v := range s.ThresholdLevels {
		if loss < v {
			return i
		}
	}
	return 0
}

func (s *AlertTriggerEvaluation) WriteLevelToFile(level int) {
	if s.FilePath != "" {
		os.WriteFile(s.FilePath, []byte(strconv.Itoa(level)), 0777)
		log.Println("WriteLevelToFile : ", strconv.Itoa(level), " to ", s.FilePath)
	}
}

func (s *AlertTriggerEvaluation) ReadFromFile() int {
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
func (s *AlertTriggerEvaluation) ShouldSend() bool {
	if s.CurrentLoss < s.ThresholdLevels[0] {
		s.ThresholdIndex = 0
		s.WriteLevelToFile(0)
		log.Println("Loss < 20 :", s.CurrentLoss, "Index:", s.ThresholdIndex)
		return false
	}
	// adjust threshold up, include index = 0
	if s.CurrentLoss > s.ThresholdLevels[s.ThresholdIndex] {
		s.ThresholdIndex = s.UpLevel(s.CurrentLoss)
		s.ThresholdAsc = true
		s.WriteLevelToFile(s.ThresholdIndex)
		log.Println("ThresholdLevel Up To :", s.ThresholdIndex, " Loss:", s.CurrentLoss, " ShouldSend", true)
		return true
	}
	// adjust threshold down, index = 0 does not need to down level
	if s.CurrentLoss < s.ThresholdLevels[s.ThresholdIndex-1] {
		s.ThresholdIndex = s.UpLevel(s.CurrentLoss)
		s.ThresholdAsc = false
		s.WriteLevelToFile(s.ThresholdIndex)
		log.Println("ThresholdLevel Down To :", s.ThresholdIndex, " Loss:", s.CurrentLoss, " ShouldSend", true)
		return true
	}
	log.Println("ThresholdLevel NOT change. Loss:", s.CurrentLoss, "Index:", s.ThresholdIndex)
	return false
}

func AlertSend(cluster Cluster, globalStat *GlobalStatistic, globalErrorStatistic map[string]int, threadhold float64) {
	payload := SlackPayload{}
	payload.AlertPayload(cluster, globalStat, globalErrorStatistic, threadhold)
	err := SlackSend(config.SlackReport.SlackAlert.WebHook, &payload)
	if err != nil {
		log.Println("SlackSend Error:", err)
	}
}
