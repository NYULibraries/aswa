package main

import (
	"errors"
	"fmt"
	a "github.com/NYULibraries/aswa/lib/application"
	"log"
	"os"
	"strings"
	"time"
)

// postTestResult posts the result of the given test to Slack.
func postTestResult(test *a.Application, channelProdId, channelDevId, token string) error {
	appStatus := test.GetStatus()
	log.Println(appStatus)

	// check if the status is not successful
	if !appStatus.Success {
		// Determine the channel to post the message
		var targetChannel string
		if strings.HasPrefix(strings.ToLower(test.Name), DEV) {
			targetChannel = channelDevId
		} else if strings.HasPrefix(strings.ToLower(test.Name), PROD) {
			targetChannel = channelProdId
		} else {
			return errors.New("app name does not start with 'dev' or 'prod'")
		}

		slackClient := NewSlackClient(token)
		if err := slackClient.PostToSlack(appStatus.String(), targetChannel); err != nil {
			return err
		}
		timestamp := time.Now().Local().Format(time.RFC1123Z)
		log.Printf("Message sent to channel %s on %s", targetChannel, timestamp)
	}

	return nil
}

func RunSyntheticTests(appData []*a.Application, channelProdId, channelDevId, token, targetAppName string) error {
	found := false // Keep track of whether the app was found in the config file
	for _, app := range appData {
		if targetAppName == "" || targetAppName == app.Name {
			found = true // The app was found in the config file
			err := postTestResult(app, channelProdId, channelDevId, token)
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

func getSlackCredentials() (string, string, string, error) {
	channelProdId := os.Getenv(envSlackChannelProdId)
	channelDevId := os.Getenv(envSlackChannelDevId)
	token := os.Getenv(envSlackToken)

	missingVars := checkMissingSlackEnvVariables(channelProdId, channelDevId, token)

	if len(missingVars) == 3 {
		log.Println("SLACK_CHANNEL_PROD_ID, SLACK_CHANNEL_DEV_ID, and SLACK_TOKEN environment variables are not set")
		return "", "", "", nil
	} else if len(missingVars) > 0 {
		errorMsg := createCustomSlackErrorMessage(missingVars)
		return "", "", "", errors.New(errorMsg)
	}

	// Check if the credentials are valid by checking auth.test
	err := ValidateSlackCredentials(token)
	if err != nil {
		return "", "", "", fmt.Errorf("invalid slack credentials: %v", err)
	}

	return channelProdId, channelDevId, token, nil
}

func checkMissingSlackEnvVariables(channelProdId, channelDevId, token string) []string {
	var missingSlackEnvVars []string

	envVars := map[string]string{
		envSlackChannelDevId:  channelDevId,
		envSlackChannelProdId: channelProdId,
		envSlackToken:         token,
	}

	for key, value := range envVars {
		if value == "" {
			missingSlackEnvVars = append(missingSlackEnvVars, key)
		}
	}

	return missingSlackEnvVars
}

func createCustomSlackErrorMessage(missingSlackEnvVars []string) string {
	errorSlackMsg := "The following environment variables must be set: "
	for i, missingSlackEnvVar := range missingSlackEnvVars {
		if i > 0 {
			errorSlackMsg += ", "
		}
		errorSlackMsg += missingSlackEnvVar
	}
	return errorSlackMsg
}

const (
	DEV                   = "dev"
	PROD                  = "prod"
	envSlackChannelDevId  = "SLACK_CHANNEL_DEV_ID"
	envSlackChannelProdId = "SLACK_CHANNEL_PROD_ID"
	envSlackToken         = "SLACK_TOKEN"
)
