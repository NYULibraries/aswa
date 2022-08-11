package main

import (
	"log"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/NYULibraries/aswa/mock_server"
	"github.com/slack-go/slack"
)

type SlackApi interface {
	PostMessage(channel string, options ...slack.MsgOption) (string, string, error)
}

type SlackService struct {
	api SlackApi
}

func NewSlackService(api SlackApi) *SlackService {
	return &SlackService{api: api}
}

func (s *SlackService) PostMessage(channel string) (string, string, error) {
	return s.api.PostMessage(channel, slack.MsgOptionText("test", false))
}

func TestPostToSlack(t *testing.T) {

	mockServer := server.New()

	client := slack.New("SLACK_TOKEN", slack.OptionAPIURL(mockServer.Server.URL+"/"))

	s := NewSlackService(client)

	channel, timestamp, err := s.PostMessage("CHANNEL_ID")

	log.Printf("Channel: %v, timestamp: %v, err: %v", channel, timestamp, err)

	assert.NoError(t, err, "should not error when posting message")
	assert.Equal(t, "CHANNEL_ID", channel, "should have posted to correct channel")
	assert.NotEmpty(t, timestamp, "should have a timestamp")

}
