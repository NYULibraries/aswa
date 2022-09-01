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

	channel := os.Getenv("SLACK_CHANNEL_ID")

	if channel == "" {
		log.Println("SLACK_CHANNEL_ID not set; aborting posting slack message!")
		return
	}

	token := os.Getenv("SLACK_TOKEN")

	if token == "" {
		log.Println("SLACK_TOKEN not set; aborting posting slack message!")
		return
	}

	//no command line args, loop through all applications and post to slack
	if len(os.Args) == 1 {
		for _, app := range appData {
			name, url, expectedStatusCode, timeout, expectedActualLocation := c.ExtractValuesFromConfig(app)

			test := a.NewApplication(name, url, expectedStatusCode, timeout, expectedActualLocation)
			appStatus := test.GetStatus()
			log.Println(appStatus)

			slackClient := NewSlackClient(token)
			slackClient.PostToSlack(appStatus.String(), channel)
		}

	} else {

		//command line arg presents, loop through all applications and post to slack if name matches
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
				err := slackClient.PostToSlack(appStatus.String(), channel)
				if err != nil {
					log.Println(err)
					panic(err)
				}
				break
			}

		}
	}
}
