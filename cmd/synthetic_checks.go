package cmd

import (
	"fmt"
	"log"
	"os"
	"time"

	a "github.com/NYULibraries/aswa/pkg/application"
	c "github.com/NYULibraries/aswa/pkg/config"
	m "github.com/NYULibraries/aswa/pkg/metrics"
)

const envOutputSlack = "OUTPUT_SLACK"

// ####################
// Synthetic Test Logic
// ####################

type FailingSyntheticTest struct {
	App       *a.Application
	AppStatus a.AppCheckStatus
}

// postTestResult constructs a string containing the result of the given test.
func postTestResult(appStatus a.AppCheckStatus) (string, error) {
	result := appStatus.String()
	timestamp := time.Now().Local().Format(time.RFC1123Z)
	log.Printf("Test result generated on %s", timestamp)

	return result, nil
}

func postToSlack(tests []FailingSyntheticTest) error {
	c.GetSlackWebhookUrl()
	c.GetClusterInfo()
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

// ###############
// Check Execution
// ###############

// DoCheck loads configuration, initializes settings, and triggers synthetic tests.
func DoCheck() error {
	yamlPath := c.GetYamlPath()
	a.SetIsPrimoVE(yamlPath)

	config, err := c.NewConfig(yamlPath)
	if err != nil {
		return err
	}

	cmdArg := c.GetCmdArg()

	return RunSyntheticTests(config.Applications, cmdArg)
}
