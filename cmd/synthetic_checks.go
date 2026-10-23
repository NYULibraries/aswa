package cmd

import (
	"fmt"
	"log"

	a "github.com/NYULibraries/aswa/pkg/application"
	c "github.com/NYULibraries/aswa/pkg/config"
	m "github.com/NYULibraries/aswa/pkg/metrics"
)

// ####################
// Synthetic Test Logic
// ####################

// RunSyntheticTests runs synthetic tests on the provided applications and pushes the results to PAG.
func RunSyntheticTests(appData []*a.Application, targetAppName string) error {
	found := false // Keep track of whether the app was found in the config file
	failing := 0

	for _, app := range appData {
		if targetAppName == "" || targetAppName == app.Name {
			found = true // The app was found in the config file
			appStatus := app.GetStatus()
			log.Println(appStatus)
			// Count every check that ran (pass or fail) so uptime has a denominator.
			m.IncrementRunCounter(app.Name)
			if !appStatus.StatusOk || !appStatus.StatusContentOk || !appStatus.StatusCSPOk {
				failing++
				m.IncrementFailedTestsCounter(app.Name)
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

	// Always push so the run-count denominator is recorded on every run,
	// even when all checks pass (PAG sums counters across pushes).
	if errorProm := m.PushMetrics(); errorProm != nil {
		log.Printf("Error encountered during metrics push: %v", errorProm)
		return errorProm // return the error to handle it accordingly
	}
	log.Printf("Success! Pushed run and failure counts for all apps to Prom-Aggregation-Gateway (%d failing)", failing)
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
