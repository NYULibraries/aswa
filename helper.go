package main

import (
	"errors"
	"fmt"
	a "github.com/NYULibraries/aswa/lib/application"
	"log"
	"os"
	"time"
)

// postTestResult posts the result of the given test to Slack.
func postTestResult(test *a.Application, channel string, token string) error {
	appStatus := test.GetStatus()
	log.Println(appStatus)

	// check if the status is not successful
	if !appStatus.Success {
		slackClient := NewSlackClient(token)
		if err := slackClient.PostToSlack(appStatus.String(), channel); err != nil {
			return err
		}
		timestamp := time.Now().Local().Format(time.RFC1123Z)
		log.Printf("Message sent to channel %s on %s", channel, timestamp)
	}

	return nil
}

func RunSyntheticTests(appData []*a.Application, channel string, token string, targetAppName string) error {
	found := false // Keep track of whether the app was found in the config file
	for _, app := range appData {

		if targetAppName == "" || targetAppName == app.Name {
			found = true // The app was found in the config file
			err := postTestResult(app, channel, token)
			if err != nil {
				return err
			}

			if targetAppName != "" {
				break
			}
		}
	}
	if !found && targetAppName != "" {
		// The app was not found in the config file
		err := fmt.Errorf("app '%s' not found in config file", targetAppName)
		log.Println(err)
		return err
	}

	return nil
}

func getSlackCredentials() (string, string, error) {
	channelId := os.Getenv(envSlackChannelId)
	token := os.Getenv(envSlackToken)
	if channelId == "" || token == "" {
		if channelId == "" && token == "" {
			// if both are not set, log a warning and return with no error
			log.Println("SLACK_CHANNEL_ID and SLACK_TOKEN environment variables are not set")
			return "", "", nil
		}
		// if only one of the variables is set, return an error
		return "", "", errors.New("SLACK_CHANNEL_ID and SLACK_TOKEN environment variables must both be set")
	}

	// Check if the credentials are valid by checking auth.test
	err := ValidateSlackCredentials(token)
	if err != nil {
		return "", "", fmt.Errorf("invalid slack credentials: %v", err)
	}

	return channelId, token, nil
}

const (
	envSlackChannelId = "SLACK_CHANNEL_ID"
	envSlackToken     = "SLACK_TOKEN"
)
