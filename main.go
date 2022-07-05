package main

import (
	"github.com/slack-go/slack"
	"log"
	"os"

	a "github.com/NYULibraries/aswa/lib/application"
	c "github.com/NYULibraries/aswa/lib/config"
)

const yamlPath = "./config/applications.yml"

func main() {

	if len(os.Args) == 1 {
		panic("No application name provided")
	}

	cmdArg := os.Args[1] // get the command line argument

	inputData, err := c.NewConfig(yamlPath)

	if err != nil {
		log.Println("Could not load config file; aborting!")
		panic(err)
	}

	appData := inputData.Applications

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

			api := slack.New(os.Getenv("SLACK_TOKEN"))

			channelID, timestamp, err := api.PostMessage(os.Getenv("CHANNEL_ID"), slack.MsgOptionText(appStatus.String(), false))

			if err != nil {
				log.Fatal(err)
				return
			}

			log.Printf("Message sent to channel %s at %s", channelID, timestamp)
			break
		}

	}
}
