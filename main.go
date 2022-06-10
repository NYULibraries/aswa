package main

import (
	"log"
	"os"
	"time"

	"github.com/NYULibraries/aswa/lib/config"
)

const yamlPath = "./config/applications.yml"

func extractValuesFromConfig(app *config.Application) (name string, url string, expectedStatusCode int, timeout time.Duration, expectedActualLocation string) {
	name = app.Name
	url = app.URL
	expectedStatusCode = app.ExpectedStatusCode
	timeout = app.Timeout
	expectedActualLocation = app.ExpectedLocation
	return
}

func containApp(applications []*config.Application, e string) bool {
	for _, application := range applications {
		if application.Name == e {
			return true
		}
	}
	return false
}

func main() {
	cmdArg := os.Args[1]

	applications, err := config.NewConfig(yamlPath)
	if err != nil {
		log.Println("Could not load config file; aborting!")
		panic(err)
	}

	if !containApp(applications.Applications, cmdArg) {
		log.Println("Application not found in config file; aborting!")
		panic(err)
	}

	for _, app := range applications.Applications {
		name, url, expectedStatusCode, timeout, expectedActualLocation := extractValuesFromConfig(app)
		if cmdArg == name {
			test := NewApplication(url, expectedStatusCode, timeout, expectedActualLocation)
			appStatus := test.GetStatus()
			log.Println(appStatus)
		}
	}
}
