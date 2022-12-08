package main

import (
	"errors"
	"fmt"
	c "github.com/NYULibraries/aswa/lib/config"
	"log"
	"os"

	a "github.com/NYULibraries/aswa/lib/application"
)

func postTestResult(test *a.Application, channel string, token string) error {
	appStatus := test.GetStatus()
	log.Println(appStatus)

	slackClient := NewSlackClient(token)
	err := slackClient.PostToSlack(appStatus.String(), channel)
	if err != nil {
		log.Println(err)
		return err
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
				log.Println(err)
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

func checkSlackEnvs() (string, string, error) {
	if os.Getenv("SLACK_CHANNEL_ID") == "" {
		err := errors.New("SLACK_CHANNEL_ID environment variable is not set")
		log.Println("Error checking Slack environment variable SLACK_CHANNEL_ID:", err)
	}
	if os.Getenv("SLACK_TOKEN") == "" {
		err := errors.New("SLACK_TOKEN environment variable is not set")
		log.Println("Error checking Slack environment variable SLACK_TOKEN:", err)
	}
	return os.Getenv("SLACK_CHANNEL_ID"), os.Getenv("SLACK_TOKEN"), nil
}
