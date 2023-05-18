package main

// This code has been refactored to improve readability, maintainability, and adhere to best practices.
// The following changes were made:
// 1. The main function has been simplified by extracting logic into separate helper functions.
// 2. The Check struct was introduced to encapsulate the main logic and associated state (e.g., the logger).
// 3. The Do() method was added to the Check struct to provide a single entry point for the main logic.
// 4. Package-level variables were used for shared instances, like the Check instance with its logger.
// 5. Constants were grouped together, and functions were ordered consistently for easier navigation.

import (
	"errors"
	"fmt"
	a "github.com/NYULibraries/aswa/lib/application"
	c "github.com/NYULibraries/aswa/lib/config"
	"log"
	"os"
	"time"
)

// Constants for environment variables
const (
	envSlackChannelId = "SLACK_CHANNEL_ID"
	envSlackToken     = "SLACK_TOKEN"
	envYamlPath       = "YAML_PATH"
)

// ###########################
// Check Struct & Initialization
// ###########################

// Check struct encapsulates the main logic and associated state (e.g., the logger)
// for running synthetic tests and posting results to Slack.
type Check struct {
	Logger *log.Logger
}

// Package-level variable 'check' holds the Check instance, initialized in the init function.
var check *Check

// init function initializes the Check instance with a logger that outputs to stdout.
func init() {
	logger := log.New(os.Stdout, "", 0)
	check = &Check{Logger: logger}
}

// ########################
// Environment & Arguments
// ########################

// getYamlPath retrieves the YAML path from the environment variable.
func getYamlPath(logger *log.Logger) string {
	yamlPath := os.Getenv(envYamlPath)
	if yamlPath == "" {
		logger.Println("WARNING: Environment variable YAML_PATH is not set")
	}
	return yamlPath
}

// getCmdArg retrieves the command line argument without using the flag package.
func getCmdArg() string {
	if len(os.Args) == 1 {
		return ""
	}
	return os.Args[1]
}

// ####################
// Synthetic Test Logic
// ####################

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

// RunSyntheticTests runs synthetic tests on the provided applications and posts results to Slack.
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

// ##########################
// Slack Credentials & Auth
// ##########################

// getSlackCredentials retrieves Slack credentials from environment variables.
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

// ###############
// Check Execution
// ###############

// Do method on Check struct.
func (ch *Check) Do() error {
	yamlPath := getYamlPath(ch.Logger)

	inputData, err := c.NewConfig(yamlPath)
	if err != nil {
		return err
	}

	appData := inputData.Applications

	channelId, token, err := getSlackCredentials()
	if err != nil {
		return err
	}

	cmdArg := getCmdArg()

	return RunSyntheticTests(appData, channelId, token, cmdArg)
}
