package main

import (
	"errors"
	"fmt"
	a "github.com/NYULibraries/aswa/lib/application"
	c "github.com/NYULibraries/aswa/lib/config"
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
		timestamp := time.Now().Local().Format(time.ANSIC)
		log.Printf("Message sent to channel %s on %s", channel, timestamp)
	}

	return nil
}

func RunTests(appData []*c.Application, channel string, token string, cmdArg string) error {
	found := false // Keep track of whether the app was found in the config file
	for _, app := range appData {
		name, url, expectedStatusCode, timeout, expectedActualLocation := c.ExtractValuesFromConfig(app)

		if cmdArg == "" || cmdArg == name {
			found = true // The app was found in the config file
			test := a.NewApplication(name, url, expectedStatusCode, timeout, expectedActualLocation)
			err := postTestResult(test, channel, token)
			if err != nil {
				return err
			}

			if cmdArg != "" {
				break
			}
		}
	}
	if !found && cmdArg != "" {
		// The app was not found in the config file
		err := fmt.Errorf("app '%s' not found in config file", cmdArg)
		log.Println(err)
		return err
	}

	return nil
}

func getSlackCredentials() (string, string, error) {
	if os.Getenv(envSlackChannelId) == "" {
		err := errors.New("SLACK_CHANNEL_ID environment variable is not set")
		log.Println("Error checking Slack environment variable SLACK_CHANNEL_ID:", err)
	}
	if os.Getenv(envSlackToken) == "" {
		err := errors.New("SLACK_TOKEN environment variable is not set")
		log.Println("Error checking Slack environment variable SLACK_TOKEN:", err)
	}
	return os.Getenv(envSlackChannelId), os.Getenv(envSlackToken), nil
}

const (
	envSlackChannelId = "SLACK_CHANNEL_ID"
	envSlackToken     = "SLACK_TOKEN"
)
