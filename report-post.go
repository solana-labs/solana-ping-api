package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/parnurzeal/gorequest"
)

// SlackSend send the payload to then Slack Channel
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

// DiscordSend send the payload to discord channel by webhook
func DiscordSend(webhookURL string, payload *DiscordPayload) []error {
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
		return []error{fmt.Errorf("Discord sending msg. Status: %v", resp.Status)}
	}
	log.Println("Discord send successfully! =>", webhookURL)
	return nil
}
