package main

import (
	"log"
	"os"

	a "github.com/NYULibraries/aswa/lib/application"
	c "github.com/NYULibraries/aswa/lib/config"
)

const yamlPath = "./config/applications.yml"

func main() {

	inputData, err := c.NewConfig(yamlPath)

	if err != nil {
		log.Println("Could not load config file; aborting!")
		panic(err)
	}

	appData := inputData.Applications

	token := os.Getenv("SLACK_TOKEN")

	if token == "" {
		log.Println("SLACK_TOKEN not set; aborting posting slack message!")
		return
	}

	if len(os.Args) == 1 {
		for _, app := range appData {
			name, url, expectedStatusCode, timeout, expectedActualLocation := c.ExtractValuesFromConfig(app)

			test := a.NewApplication(name, url, expectedStatusCode, timeout, expectedActualLocation)
			appStatus := test.GetStatus()
			log.Println(appStatus)

			slackClient := NewSlackClient(token)
			slackClient.PostToSlack(appStatus.String())
		}

	} else {

		cmdArg := os.Args[1] // get the command line argument

		if !c.ContainApp(appData, cmdArg) {
			log.Println("Application '", cmdArg, "' not found in config file; aborting!")
			panic(err)
		}

		for _, app := range appData {
			name, url, expectedStatusCode, timeout, expectedActualLocation := c.ExtractValuesFromConfig(app)

			if cmdArg == name {
				test := a.NewApplication(name, url, expectedStatusCode, timeout, expectedActualLocation)
				appStatus := test.GetStatus()
				log.Println(appStatus)

				slackClient := NewSlackClient(token)
				slackClient.PostToSlack(appStatus.String())

				break
			}

		}
	}
}
