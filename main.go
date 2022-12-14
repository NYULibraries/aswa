package main

import (
	"log"
	"os"

	c "github.com/NYULibraries/aswa/lib/config"
)

func main() {
	logger := log.New(os.Stdout, "", 0)

	inputData, err := c.NewConfig("./config/prod.applications.yml")
	if err != nil {
		logger.Fatal("Could not load config file; aborting!", err)
	}

	appData := inputData.Applications

	channel, token, err := getSlackCredentials()
	if err != nil {
		logger.Fatal("Error checking Slack environment variables:", err)
	}
	// Get the command line argument without using the flag package
	var cmdArg string
	if len(os.Args) == 1 {
		cmdArg = ""
	} else {
		cmdArg = os.Args[1]
	}

	err = RunTests(appData, channel, token, cmdArg)
	if err != nil {
		logger.Fatal("Error running tests:", err)
	}
}
