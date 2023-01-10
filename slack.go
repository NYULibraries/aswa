package main

import (
	"github.com/slack-go/slack"
	"log"
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
	// Use the `api` object to post a message to the specified Slack channel.
	_, _, err := s.api.PostMessage(channel, slack.MsgOptionText(status, false))

	// If an error occurred, return it.
	if err != nil {
		log.Println("Error posting message to Slack!!:", err)
		return err
	}

	return nil
}
