package main

import (
	"testing"

	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
)

func TestNewSlackClient(t *testing.T) {
	token := "test"
	slackClient := NewSlackClient(token)

	if assert.NotNil(t, slackClient) {
		assert.NotNil(t, slackClient.api)
	}

	if assert.NotNil(t, slackClient.api) {
		assert.Equal(t, slack.New(token), slackClient.api)
	}

}

type mockPostMessageClient struct {
	mockChannelID string
	mockStatus    string
	mockError     error
}

func (m *mockPostMessageClient) PostMessage(channel string, options ...slack.MsgOption) (string, string, error) {
	return m.mockChannelID, m.mockStatus, m.mockError
}

func testPostToSlackFunc(t *testing.T, channelID string, status string, error error) {
	mockClient := &mockPostMessageClient{
		mockChannelID: channelID,
		mockStatus:    status,
		mockError:     error,
	}
	slackClient := &SlackClient{
		api: mockClient,
	}
	slackClient.PostToSlack(status)

	if assert.Equal(t, channelID, mockClient.mockChannelID) {
		assert.Equal(t, status, mockClient.mockStatus)
		assert.Equal(t, error, mockClient.mockError)
	}

	if assert.NotNil(t, slackClient) {
		assert.NotNil(t, slackClient.api)
	}

	if assert.NotNil(t, slackClient.api) {
		assert.Equal(t, mockClient, slackClient.api)
	}

}

func TestPostToSlack(t *testing.T) {
	var tests = []struct {
		description string
		channelID   string
		status      string
		error       error
	}{
		{"Valid channelID and status", "C1234567890", "Status", nil},
		{"Invalid channelID", "", "Status", nil},
		{"Invalid status", "C1234567890", "", nil},
		{"Invalid channelID and status", "", "", nil},
		{"Invalid channelID and status and error", "", "", error(nil)},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			testPostToSlackFunc(t, test.channelID, test.status, test.error)
		})
	}
}
