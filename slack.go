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

func (s *SlackClient) PostToSlack(status string, channel string) {

	channelID, _, err := s.api.PostMessage(channel, slack.MsgOptionText(status, false))

	if err != nil {
		//Used log.Print instead of log.Fatal.The function log.Fatal is strictly for reporting your program's final breath.
		//https://stackoverflow.com/questions/45797858/testing-log-fatalf-in-go
		log.Print(err)
	}

	timestamp := time.Now().Local().Format(time.ANSIC)

	log.Printf("Message sent to channel %s on %s", channelID, timestamp)
}
