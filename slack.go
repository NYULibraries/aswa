package main

import (
	"github.com/slack-go/slack"
	"log"
	"time"
)

type postMessageClient interface {
	PostMessage(channel string, options ...slack.MsgOption) (string, string, error)
}

type SlackClient struct {
	api postMessageClient
}

func NewSlackClient(token string) *SlackClient {
	api := slack.New(token)
	return &SlackClient{api}
}

func (s *SlackClient) PostToSlack(status string, channel string) error {

	channelID, _, err := s.api.PostMessage(channel, slack.MsgOptionText(status, false))

	if err != nil {
		return err
	}

	timestamp := time.Now().Local().Format(time.ANSIC)

	log.Printf("Message sent to channel %s on %s", channelID, timestamp)
	return nil
}
