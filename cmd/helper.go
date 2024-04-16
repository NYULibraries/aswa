package main

import (
	"fmt"
	a "github.com/NYULibraries/aswa/pkg/application"
	c "github.com/NYULibraries/aswa/pkg/config"
	m "github.com/NYULibraries/aswa/pkg/metrics"
	"log"
	"os"
	"time"
)

// Constants for environment variables
const (
	envClusterInfo     = "CLUSTER_INFO"
	envOutputSlack     = "OUTPUT_SLACK"
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
		logger.Println("Environment variable for YAML path not found, using default")
		yamlPath = "config/dev.applications.yml"
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

func postToSlack(tests []FailingSyntheticTest) error {
	getSlackWebhookUrl()
	getClusterInfo()
	for _, test := range tests {
		result, err := postTestResult(test.AppStatus)
		if err != nil {
			return err
		}
		fmt.Println(result)
	}
	return nil
}

// RunSyntheticTests runs synthetic tests on the provided applications and posts results to Slack.
func RunSyntheticTests(appData []*a.Application, targetAppName string) error {
	found := false // Keep track of whether the app was found in the config file

	var failingSyntheticTests []FailingSyntheticTest

	IsOutputSlack := os.Getenv(envOutputSlack) == "true"

	for _, app := range appData {
		if targetAppName == "" || targetAppName == app.Name {
			found = true // The app was found in the config file
			appStatus := app.GetStatus()
			log.Println(appStatus)
			if !appStatus.StatusOk || !appStatus.StatusContentOk || !appStatus.StatusCSPOk {
				failingSyntheticTests = append(failingSyntheticTests, FailingSyntheticTest{App: app, AppStatus: *appStatus})
				if !IsOutputSlack {
					m.IncrementFailedTestsCounter(app.Name)
				}
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

	if len(failingSyntheticTests) > 0 {
		if IsOutputSlack {
			return postToSlack(failingSyntheticTests)
		}
		// Push metrics to Prom-Aggregation-Gateway if OUTPUT_SLACK is not set
		if errorProm := m.PushMetrics(); errorProm != nil {
			log.Printf("Error encountered during metrics push: %v", errorProm)
			return errorProm // return the error to handle it accordingly
		}
		log.Println("Success! Pushed failed test count for all apps to Prom-Aggregation-Gateway")
	} else {
		log.Println("No failed tests. No actions taken.")
	}
	return nil
}

// #################################
// Slack WebHook Url & Cluster Info
// #################################

// getSlackCredentials retrieves Slack credentials from environment variables.
func getSlackWebhookUrl() string {
	slackWebhookUrl := os.Getenv(envSlackWebhookUrl)
	if slackWebhookUrl == "" {
		log.Println("SLACK_WEBHOOK_URL is not set")
		return ""
	}
	return slackWebhookUrl
}

// getClusterInfo retrieves the cluster info from environment variables.
func getClusterInfo() string {
	clusterInfo := os.Getenv(envClusterInfo)
	if clusterInfo == "" {
		log.Println("CLUSTER_INFO is not set")
		return ""
	}
	return clusterInfo
}

// ###############
// Check Execution
// ###############

// Do method on Check struct.
func (ch *Check) Do() error {
	yamlPath := getYamlPath(ch.Logger)
	a.SetIsPrimoVE(yamlPath)

	inputData, err := c.NewConfig(yamlPath)
	if err != nil {
		return err
	}

	appData := inputData.Applications

	cmdArg := getCmdArg()

	return RunSyntheticTests(appData, cmdArg)
}
