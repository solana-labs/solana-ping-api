package main

import (
	"encoding/json"
	"fmt"

	"github.com/parnurzeal/gorequest"
)

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
		log.Info(resp.StatusCode, " payload:", string(data))
		return []error{fmt.Errorf("slack sending msg. Status: %v", resp.Status)}
	}

	log.Info("Slack Send Successfully =>")
	return nil
}
