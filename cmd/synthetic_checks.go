package cmd

import (
	"fmt"
	"log"
	"os"
	"strconv"
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
	AppStatus a.AppCheckStatus
}

// postTestResult constructs a string containing the result of the given test.
func postTestResult(appStatus a.AppCheckStatus) string {
	result := appStatus.String()
	timestamp := time.Now().Local().Format(time.RFC1123Z)
	log.Printf("Test result generated on %s", timestamp)

	return result
}

func postToSlack(tests []FailingSyntheticTest) error {
	for _, test := range tests {
		fmt.Println(postTestResult(test.AppStatus))
	}
	return nil
}

// RunSyntheticTests runs synthetic tests on the provided applications and posts results to Slack.
func RunSyntheticTests(appData []*a.Application, targetAppName string) error {
	found := false // Keep track of whether the app was found in the config file

	var failingSyntheticTests []FailingSyntheticTest

	v, _ := strconv.ParseBool(os.Getenv(envOutputSlack))
	IsOutputSlack := v

	for _, app := range appData {
		if targetAppName == "" || targetAppName == app.Name {
			found = true // The app was found in the config file
			appStatus := app.GetStatus()
			log.Println(appStatus)
			// Count every check that ran (pass or fail) so uptime has a denominator.
			if !IsOutputSlack {
				m.IncrementRunCounter(app.Name)
			}
			if !appStatus.StatusOk || !appStatus.StatusContentOk || !appStatus.StatusCSPOk {
				failingSyntheticTests = append(failingSyntheticTests, FailingSyntheticTest{AppStatus: *appStatus})
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

	if IsOutputSlack {
		if len(failingSyntheticTests) > 0 {
			return postToSlack(failingSyntheticTests)
		}
		log.Println("No failed tests. No actions taken.")
		return nil
	}

	// Metrics mode: always push so the run-count denominator is recorded on every run,
	// even when all checks pass (PAG sums counters across pushes).
	if errorProm := m.PushMetrics(); errorProm != nil {
		log.Printf("Error encountered during metrics push: %v", errorProm)
		return errorProm // return the error to handle it accordingly
	}
	log.Printf("Success! Pushed run and failure counts for all apps to Prom-Aggregation-Gateway (%d failing)", len(failingSyntheticTests))
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
