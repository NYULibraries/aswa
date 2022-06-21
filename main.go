package main

import (
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

	CmdArg := os.Args[1] // get the command line argument

	inputData, err := c.NewConfig(yamlPath)

	appData := inputData.Applications

	if err != nil {
		log.Println("Could not load config file; aborting!")
		panic(err)
	}

	if !c.ContainApp(appData, CmdArg) {
		log.Println("Application '", CmdArg, "' not found in config file; aborting!")
		panic(err)
	}

	for _, app := range appData {
		name, url, expectedStatusCode, timeout, expectedActualLocation := c.ExtractValuesFromConfig(app)

		if os.Args[1] == name {
			test := a.NewApplication(name, url, expectedStatusCode, timeout, expectedActualLocation)
			appStatus := test.GetStatus()
			log.Println(appStatus)
		}

	}
}
