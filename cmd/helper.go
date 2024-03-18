package main

import (
	"bytes"
	"fmt"
	a "github.com/NYULibraries/aswa/pkg/application"
	c "github.com/NYULibraries/aswa/pkg/config"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"io"
	"net/http"
	"strings"

	"log"
	"os"
	"time"
)

// Constants for environment variables
const (
	envClusterInfo     = "CLUSTER_INFO"
	envPushgatewayUrl  = "PUSHGATEWAY_URL"
	envSlackWebhookUrl = "SLACK_WEBHOOK_URL"
	envYamlPath        = "YAML_PATH"
)

// ###############################################################
// Define a Prometheus counter to count the number of failed tests
// ###############################################################

var (
	failedTests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "aswa_failed_tests_total",
			Help: "Total number of failed synthetic tests.",
		},
		[]string{"app"},
	)
)

// init function to increment the failed tests counter
func incrementFailedTestsCounter(appName string) {
	failedTests.With(prometheus.Labels{"app": appName}).Inc()
}

// Create a struct that implements the interface using the push package
type PrometheusPusher struct {
	url       string
	namespace string
	collector prometheus.Collector
	appName   string
	counter   int
}

// Push sends a Prometheus counter metric for a specific application to a Pushgateway.
// It formats the metric in Prometheus exposition format, constructs a POST request to the Pushgateway,
// and handles the response. Errors during metric serialization, request creation, or server response
// are returned to the caller.
func (p *PrometheusPusher) Push(appName string, counter prometheus.Counter) error {
	metricFamily := &dto.Metric{}
	if err := counter.Write(metricFamily); err != nil {
		log.Fatalf("Could not write Metric: %v", err)
	}

	// Convert the metric to the Prometheus exposition format
	metricData := fmt.Sprintf("%s{%s=\"%s\"} %f\n", p.namespace+"FailedTests", "app", appName, *metricFamily.Counter.Value)

	// Create a new HTTP request
	req, err := http.NewRequest("POST", p.url+"/metrics/job/"+p.namespace, strings.NewReader(metricData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "text/plain")

	// Send the request using the http.DefaultClient
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	// Check the response
	if resp.StatusCode != http.StatusOK {
		buf := new(bytes.Buffer)
		_, err = io.Copy(buf, resp.Body)
		return fmt.Errorf("response Status: %s, Response Body: %s", resp.Status, buf.String())
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

// getPushgatewayUrl retrieves the pushgateway url from environment variables.
func getPushgatewayUrl() string {
	pushgatewayUrl := os.Getenv(envPushgatewayUrl)
	if pushgatewayUrl == "" {
		log.Println("PUSHGATEWAY_URL is not set")
		return ""
	}
	return pushgatewayUrl
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
