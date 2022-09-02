package main

import (
	"log"
	"os"

	a "github.com/NYULibraries/aswa/lib/application"
	c "github.com/NYULibraries/aswa/lib/config"
)

func postTestResult(test *a.Application, channel string, token string) error {
	appStatus := test.GetStatus()
	log.Println(appStatus)

	slackClient := NewSlackClient(token)
	err := slackClient.PostToSlack(appStatus.String(), channel)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func RunTestsNoCmdArgs(appData []*c.Application, channel string, token string) error {
	if len(os.Args) == 1 {
		for _, app := range appData {
			name, url, expectedStatusCode, timeout, expectedActualLocation := c.ExtractValuesFromConfig(app)

			test := a.NewApplication(name, url, expectedStatusCode, timeout, expectedActualLocation)
			err := postTestResult(test, channel, token)
			if err != nil {
				log.Println(err)
				return err
			}
		}

	}
	return nil
}

func RunTests(appData []*c.Application, channel string, token string, cmdArg string) error {

	if !c.ContainApp(appData, cmdArg) {
		log.Println("Application '", cmdArg, "' not found in config file; aborting!")
	}

	for _, app := range appData {
		name, url, expectedStatusCode, timeout, expectedActualLocation := c.ExtractValuesFromConfig(app)

		if cmdArg == name {
			test := a.NewApplication(name, url, expectedStatusCode, timeout, expectedActualLocation)
			err := postTestResult(test, channel, token)
			if err != nil {
				log.Println(err)
				return err
			}
			break
		}

	}
	return nil
}
