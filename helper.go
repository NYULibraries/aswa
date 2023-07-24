package main

// This code has been refactored to improve readability, maintainability, and adhere to best practices.
// The following changes were made:
// 1. The main function has been simplified by extracting logic into separate helper functions.
// 2. The Check struct was introduced to encapsulate the main logic and associated state (e.g., the logger).
// 3. The Do() method was added to the Check struct to provide a single entry point for the main logic.
// 4. Package-level variables were used for shared instances, like the Check instance with its logger.
// 5. Constants were grouped together, and functions were ordered consistently for easier navigation.

import (
	"fmt"
	a "github.com/NYULibraries/aswa/lib/application"
	c "github.com/NYULibraries/aswa/lib/config"
	"log"
	"os"
	"time"
)

// Constants for environment variables
const (
	envClusterInfo     = "CLUSTER_INFO"
	envSlackWebhookUrl = "SLACK_WEBHOOK_URL"
	envYamlPath        = "YAML_PATH"
)

// #############################
// Check Struct & Initialization
// #############################

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

// postTestResult constructs a string containing the result of the given test.
func postTestResult(appStatus a.ApplicationStatus) (string, error) {
	result := appStatus.String()
	timestamp := time.Now().Local().Format(time.RFC1123Z)
	log.Printf("Test result generated on %s", timestamp)

	return result, nil
}

// RunSyntheticTests runs synthetic tests on the provided applications and posts results to Slack.
func RunSyntheticTests(appData []*a.Application, targetAppName string) error {
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
		result, err := postTestResult(failingTest.AppStatus)
		if err != nil {
			return err
		}
		fmt.Println(result)
	}
	return nil
}

// ################################
// Slack WebHook Url & Cluster Info
// ################################

// getSlackCredentials retrieves Slack credentials from environment variables.
func getSlackWebhookUrl() (string, error) {
	slackWebhookUrl := os.Getenv(envSlackWebhookUrl)
	if slackWebhookUrl == "" {
		log.Println("SLACK_WEBHOOK_URL is not set")
		return "", nil
	}
	return slackWebhookUrl, nil
}

// getClusterInfo retrieves the cluster info from environment variables.
func getClusterInfo() (string, error) {
	clusterInfo := os.Getenv(envClusterInfo)
	if clusterInfo == "" {
		log.Println("CLUSTER_INFO is not set")
		return "", nil
	}
	return clusterInfo, nil
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

	_, err = getSlackWebhookUrl()
	if err != nil {
		return err
	}

	_, err = getClusterInfo()
	if err != nil {
		return err
	}

	cmdArg := getCmdArg()

	return RunSyntheticTests(appData, cmdArg)
}
