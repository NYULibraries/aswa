package main

import (
	"log"
	"time"

	"github.com/NYULibraries/aswa/lib/config"
)

const yamlPath = "./config/applications.yml"

func extractValuesFromConfig(app *config.Application) (url string, expectedStatusCode int, timeout time.Duration, expectedActualLocation string) {
	url = app.URL
	expectedStatusCode = app.ExpectedStatusCode
	timeout = app.Timeout
	expectedActualLocation = app.ExpectedLocation
	return
}

func main() {

	applications, err := config.NewConfig(yamlPath)
	if err != nil {
		log.Println("Could not load config file; aborting!")
		panic(err)
	}

	for _, app := range applications.Applications {
		url, expectedStatusCode, timeout, expectedActualLocation := extractValuesFromConfig(app)

		test := NewApplication(url, expectedStatusCode, timeout, expectedActualLocation)
		appStatus := test.GetStatus()
		log.Println(appStatus)
	}
}
