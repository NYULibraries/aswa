package main

import (
	"errors"
	"fmt"
	a "github.com/NYULibraries/aswa/lib/application"
	"log"
	"os"
	"time"
)

type FailingSyntheticTest struct {
	App       *a.Application
	AppStatus a.ApplicationStatus
}

// postTestResult posts the result of the given test to Slack.
func postTestResult(appStatus a.ApplicationStatus, channel string, token string) error {

	slackClient := NewSlackClient(token)
	if err := slackClient.PostToSlack(appStatus.String(), channel); err != nil {
		return err
	}
	timestamp := time.Now().Local().Format(time.RFC1123Z)
	log.Printf("Message sent to %s channel on %s", channel, timestamp)

	return nil
}

func RunSyntheticTests(appData []*a.Application, channel string, token string, targetAppName string) error {
	found := false // Keep track of whether the app was found in the config file

	var failingSyntheticTests []FailingSyntheticTest

	for _, app := range appData {
		if targetAppName == "" || targetAppName == app.Name {
			found = true // The app was found in the config file
			appStatus := app.GetStatus()
			log.Println(appStatus)
			if !appStatus.StatusOk || !appStatus.StatusContentOk {
				failingSyntheticTests = append(failingSyntheticTests, FailingSyntheticTest{App: app, AppStatus: *appStatus})
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

	// Post failing test results after running tests on all applications
	for _, failingTest := range failingSyntheticTests {
		err := postTestResult(failingTest.AppStatus, channel, token)
		if err != nil {
			return err
		}
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
