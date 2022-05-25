package main

import (
	"log"
	"os"
	"strconv"
	"time"
)

const defaultTimeout = 1 * time.Minute

func loadApplicationVariables() (string, int, time.Duration, string) {
	url := os.Getenv("ASWA_URL")
	expectedStatusCode, err := strconv.Atoi(os.Getenv("ASWA_EXPECTED_STATUS"))
	if err != nil {
		log.Println("Could not parse expected status; aborting!")
		panic(err)
	}

	timeout, err := time.ParseDuration(os.Getenv("ASWA_TIMEOUT"))
	if timeout <= 0 || err != nil {
		log.Println("No valid timeout provided; using default")
		timeout = defaultTimeout
	}

	expectedActualLocation := os.Getenv("ASWA_EXPECTED_LOCATION")

	return url, expectedStatusCode, timeout, expectedActualLocation
}

func main() {
	url, expectedStatusCode, timeout, expectedActualLocation := loadApplicationVariables()

	test := NewApplication(url, expectedStatusCode, timeout, expectedActualLocation)
	appStatus := test.GetStatus()
	log.Println(appStatus)
}
