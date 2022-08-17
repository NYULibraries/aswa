package main

import (
	"github.com/slack-go/slack"
	"log"
	"os"
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

func (s *SlackClient) PostToSlack(status string) {
	channel := os.Getenv("SLACK_CHANNEL_ID")

	if channel == "" {
		log.Println("SLACK_CHANNEL_ID not set; aborting posting slack message!")
		return
	}

	channelID, _, err := s.api.PostMessage(channel, slack.MsgOptionText(status, false))

	if err != nil {
		log.Fatal(err)
		return
	}

	timestamp := time.Now().Local().Format(time.ANSIC)

	log.Printf("Message sent to channel %s on %s", channelID, timestamp)
}
