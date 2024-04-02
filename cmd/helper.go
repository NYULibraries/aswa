package main

import (
	"fmt"
	a "github.com/NYULibraries/aswa/pkg/application"
	c "github.com/NYULibraries/aswa/pkg/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
	"log"
	"os"
)

// Constants for environment variables
const (
	envClusterInfo               = "CLUSTER_INFO"
	envPromAggregationGatewayUrl = "PROM_AGGREGATION_GATEWAY_URL"
	envSlackWebhookUrl           = "SLACK_WEBHOOK_URL"
	envYamlPath                  = "YAML_PATH"
)

// ###############################################################
// Define a Prometheus counter to count the number of failed tests
// ###############################################################

var (
	failedTests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "aswa_checks",
			Help: "Failed synthetic test.",
		},
		[]string{"app"},
	)
)

// init function to increment the failed tests counter
func incrementFailedTestsCounter(appName string) {
	failedTests.With(prometheus.Labels{"app": appName}).Inc()
}

// PushMetrics pushes all collected metrics to the pag.
func PushMetrics() error {
	pusher := push.New(getPromAggregationgatewayUrl(), "monitoring")
	if err := pusher.Collector(failedTests).Push(); err != nil {
		return err
	}
	return nil
}

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
	prometheus.MustRegister(failedTests)
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

// RunSyntheticTests runs synthetic tests on the provided applications and posts results to Slack.
func RunSyntheticTests(appData []*a.Application, targetAppName string) error {
	found := false // Keep track of whether the app was found in the config file

	var failingSyntheticTests []FailingSyntheticTest

	for _, app := range appData {
		if targetAppName == "" || targetAppName == app.Name {
			found = true // The app was found in the config file
			appStatus := app.GetStatus()
			log.Println(appStatus)
			if !appStatus.StatusOk || !appStatus.StatusContentOk || !appStatus.StatusCSPOk {
				failingSyntheticTests = append(failingSyntheticTests, FailingSyntheticTest{App: app, AppStatus: *appStatus})
				incrementFailedTestsCounter(app.Name)
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

	// Push metrics after tests are run for all applications
	if errorProm := PushMetrics(); errorProm != nil {
		log.Printf("Error encountered during metrics push: %v", errorProm)
		return errorProm // return the error to handle it accordingly
	}
	log.Println("Success! Pushed failed test count for all apps to Prom-Aggregation-Gateway")

	return nil
}

// ##################################################
// Slack WebHook Url & Cluster Info & Prom-Aggregation-Gateway Url
// ##################################################

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

// getPromAggregationgatewayUrl retrieves the pag url from environment variables.
func getPromAggregationgatewayUrl() string {
	promAggregationGatewayUrl := os.Getenv(envPromAggregationGatewayUrl)
	if promAggregationGatewayUrl == "" {
		log.Println("PROM_AGGREGATION_GATEWAY_URL is not set")
		return ""
	}
	return promAggregationGatewayUrl
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

	getSlackWebhookUrl()
	getClusterInfo()

	cmdArg := getCmdArg()

	return RunSyntheticTests(appData, cmdArg)
}
