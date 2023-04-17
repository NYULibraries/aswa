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

type FailingSyntheticTest struct {
	App       *a.Application
	AppStatus a.ApplicationStatus
}

// postTestResult posts the result of the given test to Slack.
func postTestResult(test *a.Application, appStatus a.ApplicationStatus, channelDevId, channelProdId, channelSaasId, token string) error {
	// Determine the channel to post the message
	var targetChannel string
	var slackChannelToPost string
	if strings.HasPrefix(strings.ToLower(test.Name), DEV) {
		targetChannel = channelDevId
		slackChannelToPost = DevChannel
	} else if strings.HasPrefix(strings.ToLower(test.Name), PROD) {
		targetChannel = channelProdId
		slackChannelToPost = ProdChannel
	} else if strings.HasPrefix(strings.ToLower(test.Name), SAAS) {
		targetChannel = channelSaasId
		slackChannelToPost = SaasChannel
	} else {
		return errors.New("app name does not start with 'dev' or 'prod'")
	}

	slackClient := NewSlackClient(token)
	if err := slackClient.PostToSlack(appStatus.String(), targetChannel); err != nil {
		return err
	}
	timestamp := time.Now().Local().Format(time.RFC1123Z)
	log.Printf("Message sent to %s channel on %s", slackChannelToPost, timestamp)

	return nil
}

func RunSyntheticTests(appData []*a.Application, channelDevId, channelProdId, channelSaasId, token, targetAppName string) error {
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
		err := postTestResult(failingTest.App, failingTest.AppStatus, channelDevId, channelProdId, channelSaasId, token)
		if err != nil {
			return err
		}
	}
	return nil
}

func getSlackCredentials() (string, string, string, string, error) {
	channelDevId := os.Getenv(envSlackChannelDevId)
	channelProdId := os.Getenv(envSlackChannelProdId)
	channelSaasId := os.Getenv(envSlackChannelSaasId)
	token := os.Getenv(envSlackToken)

	missingVars := checkMissingSlackEnvVariables(channelDevId, channelProdId, channelSaasId, token)

	if len(missingVars) == 4 {
		log.Println("SLACK_CHANNEL_DEV_ID, SLACK_CHANNEL_PROD_ID, SLACK_CHANNEL_SAAS_ID and SLACK_TOKEN environment variables are not set")
		return "", "", "", "", nil
	} else if len(missingVars) > 0 {
		errorMsg := createCustomSlackErrorMessage(missingVars)
		return "", "", "", "", errors.New(errorMsg)
	}

	// Check if the credentials are valid by checking auth.test
	err := ValidateSlackCredentials(token)
	if err != nil {
		return "", "", "", "", fmt.Errorf("invalid slack credentials: %v", err)
	}

	return channelDevId, channelProdId, channelSaasId, token, nil
}

func checkMissingSlackEnvVariables(channelDevId, channelProdId, channelSaasId, token string) []string {
	var missingSlackEnvVars []string

	envVars := map[string]string{
		envSlackChannelDevId:  channelDevId,
		envSlackChannelProdId: channelProdId,
		envSlackChannelSaasId: channelSaasId,
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
	DEV  = "dev"
	PROD = "prod"
	SAAS = "saas"

	envSlackChannelDevId  = "SLACK_CHANNEL_DEV_ID"
	envSlackChannelProdId = "SLACK_CHANNEL_PROD_ID"
	envSlackChannelSaasId = "SLACK_CHANNEL_SAAS_ID"
	envSlackToken         = "SLACK_TOKEN"

	DevChannel  = "synthetic-tests-dev"
	ProdChannel = "synthetic-tests-prod"
	SaasChannel = "synthetic-tests-saas"
)
