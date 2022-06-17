package main

import (
	"log"
	"os"

	"github.com/NYULibraries/aswa/lib/application"
	"github.com/NYULibraries/aswa/lib/config"
)

const yamlPath = "./config/applications.yml"

func main() {

	if len(os.Args) == 1 {
		panic("No application name provided")
	}

	cmdArg := os.Args[1] // get the command line argument

	applications, err := config.NewConfig(yamlPath)

	if err != nil {
		log.Println("Could not load config file; aborting!")
		panic(err)
	}

	if !config.ContainApp(applications.Applications, cmdArg) {
		log.Println("Application '", cmdArg, "' not found in config file; aborting!")
		panic(err)
	}

	for _, app := range applications.Applications {
		name, url, expectedStatusCode, timeout, expectedActualLocation := config.ExtractValuesFromConfig(app)

		if cmdArg == name {
			test := application.NewApplication(name, url, expectedStatusCode, timeout, expectedActualLocation)
			appStatus := test.GetStatus()
			log.Println(appStatus)
		}

	}
}
