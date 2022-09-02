package main

import (
	"log"
	"os"

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

	channel, token, err := checkSlackEnvs()
	if err != nil {
		log.Println(err)
		panic(err)
	}

	//no command line args, loop through all applications and post to slack
	if len(os.Args) == 1 {
		error := RunTestsNoCmdArgs(appData, channel, token)
		if error != nil {
			log.Println(error)
			panic(error)
		}
	} else {
		cmdArg := os.Args[1]

		error := RunTests(appData, channel, token, cmdArg)
		if error != nil {
			log.Println(error)
			panic(error)
		}
	}
}

func checkSlackEnvs() (string, string, error) {
	channel := os.Getenv("SLACK_CHANNEL_ID")
	if channel == "" {
		log.Fatal("SLACK_CHANNEL_ID not set; aborting posting slack message!")
	}

	token := os.Getenv("SLACK_TOKEN")
	if token == "" {
		log.Fatal("SLACK_TOKEN not set; aborting posting slack message!")
	}

	return channel, token, nil
}
