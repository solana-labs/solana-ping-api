package main

import (
	"log"
	"os"
	"strconv"
	"strings"
)

const AlertTriggerNameLength = 30

type AlertTrigger struct {
	Name            string
	LastLoss        float64
	CurrentLoss     float64
	ThresholdIndex  int
	ThresholdLevels []float64
	ThresholdAsc    bool
	FilePath        string
}

func NewAlertTrigger(conf ClusterConfig) AlertTrigger {
	s := AlertTrigger{}
	s.Name = "default"
	s.FilePath = conf.Report.LevelFilePath
	s.CurrentLoss = 0
	s.LastLoss = 0
	s.ThresholdLevels = []float64{float64(conf.Report.LossThreshold), float64(50), float64(75), float64(100)}
	s.ThresholdIndex = s.ReadIndexFromFile()
	return s
}
func NewAlertTriggerByParams(alertName string, levelFilePath string, lossThreshold float64) AlertTrigger {
	s := AlertTrigger{}
	alertName = strings.Trim(alertName, " ")
	if len(alertName) > AlertTriggerNameLength {
		log.Panic("alertName length > ", AlertTriggerNameLength)
	}
	if len(alertName) == 0 {
		s.Name = "default"
	} else {
		s.Name = alertName
	}

	s.FilePath = levelFilePath
	s.CurrentLoss = 0
	s.LastLoss = 0
	s.ThresholdLevels = []float64{lossThreshold, float64(50), float64(75), float64(100)}
	s.ThresholdIndex = s.ReadIndexFromFile()
	return s
}
func (s *AlertTrigger) Update(currentLoss float64) {
	s.LastLoss = s.CurrentLoss
	s.CurrentLoss = currentLoss * 100
}

// get a threshold index which is 1 level higher than loss value
func (s *AlertTrigger) UpThresholdIndex(loss float64) int {
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

func (s *AlertTrigger) WritIndexToFile(level int) {
	if s.FilePath != "" {
		os.WriteFile(s.FilePath, []byte(strconv.Itoa(level)), 0777)
		log.Println(s.Name, " trigger ", "WriteLevelToFile : ", strconv.Itoa(level), " to ", s.FilePath)
	}
}

func (s *AlertTrigger) ReadIndexFromFile() int {
	if s.FilePath != "" {
		v, err := os.ReadFile(s.FilePath)
		if err != nil {
			return 0
		}
		level, err := strconv.Atoi(string(v))
		if err != nil {
			return 0
		}
		log.Println(s.Name, " trigger ", "ReadFromFile : ", level, " from ", s.FilePath)
		return level
	}
	return 0
}

// Doing rule here
func (s *AlertTrigger) ShouldAlertSend() bool {
	if s.ThresholdLevels[0] == 0 {
		return true
	}
	if s.CurrentLoss < s.ThresholdLevels[0] {
		s.ThresholdIndex = 0
		s.WritIndexToFile(0)
		log.Println(s.Name, " trigger ", "Loss = ", s.CurrentLoss, " < ", s.ThresholdLevels[0], "Index:", s.ThresholdIndex)
		return false
	}
	// adjust threshold up, include index = 0
	if s.CurrentLoss > s.ThresholdLevels[s.ThresholdIndex] {
		s.ThresholdIndex = s.UpThresholdIndex(s.CurrentLoss)
		s.ThresholdAsc = true
		s.WritIndexToFile(s.ThresholdIndex)
		log.Println(s.Name, " trigger ", "ThresholdLevel Up To :", s.ThresholdIndex, " Loss:", s.CurrentLoss, " ShouldSend", true)
		return true
	}
	// adjust threshold down, index = 0 does not need to down level
	if s.CurrentLoss < s.ThresholdLevels[s.ThresholdIndex-1] {
		s.ThresholdIndex = s.UpThresholdIndex(s.CurrentLoss)
		s.ThresholdAsc = false
		s.WritIndexToFile(s.ThresholdIndex)
		log.Println(s.Name, " trigger ", "ThresholdLevel Down To :", s.ThresholdIndex, " Loss:", s.CurrentLoss, " ShouldSend", true)
		return true
	}
	log.Println(s.Name, " trigger ", "ThresholdLevel NOT change. Loss:", s.CurrentLoss, "Index:", s.ThresholdIndex)
	return false
}
