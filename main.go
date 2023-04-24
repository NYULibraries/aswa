package main

import (
	"log"
	"os"

	c "github.com/NYULibraries/aswa/lib/config"
)

func main() {
	logger := log.New(os.Stdout, "", 0)

	yamlPath := os.Getenv(envYamlPath)
	if yamlPath == "" {
		logger.Println("WARNING: Environment variable YAML_PATH is not set")
	}

	inputData, err := c.NewConfig(yamlPath)
	if err != nil {
		logger.Fatal("Could not load config file; aborting!", err)
	}

	appData := inputData.Applications

	var channelId, token, errorSlack = getSlackCredentials()

	if errorSlack != nil {
		logger.Fatal("Error checking Slack environment variables:", errorSlack)
	}
	// Get the command line argument without using the flag package
	var cmdArg string
	if len(os.Args) == 1 {
		cmdArg = ""
	} else {
		cmdArg = os.Args[1]
	}

	err = RunSyntheticTests(appData, channelId, token, cmdArg)
	if err != nil {
		logger.Fatal("Error running tests:", err)
	}
}

const envYamlPath = "YAML_PATH"
