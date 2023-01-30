package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/slack-go/slack"
	"io"
	"log"
	"net/http"
	"net/url"
)

type postMessageClient interface {
	PostMessage(channel string, options ...slack.MsgOption) (string, string, error)
}

type SlackClient struct {
	api postMessageClient
}

type AuthTestResponse struct {
	OK    bool `json:"ok"`
	Error any  `json:"error,omitempty"`
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

func ValidateSlackCredentials(token string) error {
	data := url.Values{}
	data.Set("token", token)

	req, err := http.NewRequest("POST", "https://slack.com/api/auth.test", bytes.NewBufferString(data.Encode()))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Println("Error closing response body:", err)
		}
	}(res.Body)

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, res.Body)
	if err != nil {
		return err
	}

	var authTestResponse AuthTestResponse
	if err := json.Unmarshal(buf.Bytes(), &authTestResponse); err != nil {
		return err
	}

	if !authTestResponse.OK {
		return fmt.Errorf("slack API error: %s", authTestResponse.Error)
	}

	return nil
}
