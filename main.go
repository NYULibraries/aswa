package main

import (
	"log"

	"github.com/NYULibraries/aswa/lib/config"
)

func main() {

	applications, err := config.NewConfig("./config/applications.yml")
	if err != nil {
		log.Println("Could not load config file; aborting!")
		panic(err)
	}

	for _, app := range applications.Applications {
		url := app.URL
		expectedStatusCode := app.ExpectedStatusCode
		timeout := app.Timeout
		expectedActualLocation := app.ExpectedLocation

		test := NewApplication(url, expectedStatusCode, timeout, expectedActualLocation)
		appStatus := test.GetStatus()
		log.Println(appStatus)
	}
}
