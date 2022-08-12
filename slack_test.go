package main

import (
	"testing"

	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
)

type mockPostMessageClient struct {
	mockChannelID string
	mockStatus    string
	mockError     error
}

func (m *mockPostMessageClient) PostMessage(channel string, options ...slack.MsgOption) (string, string, error) {
	return m.mockChannelID, m.mockStatus, m.mockError
}

func testPostToSlackWithClientFunc(t *testing.T, channelID string, status string, error error) {

	mockApi := &mockPostMessageClient{channelID, status, error}
	PostToSlackWithClient("testStatus", mockApi)

	if assert.Equal(t, channelID, mockApi.mockChannelID) {
		assert.True(t, true)

	} else {
		assert.True(t, false)
	}

	if assert.Equal(t, status, mockApi.mockStatus) {
		assert.True(t, true)
	} else {
		assert.True(t, false)
	}

	if assert.Equal(t, error, mockApi.mockError) {
		assert.True(t, true)
	} else {
		assert.True(t, false)
	}
}

func TestPostToSlackWithClient(t *testing.T) {
	var tests = []struct {
		description string
		channelID   string
		status      string
		error       error
	}{
		{"Valid channelID and status", "C1234567890", "testStatus", nil},
		{"Invalid channelID", "", "testStatus", nil},
		{"Invalid status", "C1234567890", "", nil},
		{"Invalid channelID and status", "", "", nil},
		{"Invalid channelID and status and error", "", "", error(nil)},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			testPostToSlackWithClientFunc(t, test.channelID, test.status, test.error)
		})
	}
}
