package main

import (
	"log"
	"os"
	"strconv"
	"time"
)

const defaultTimeout = 1 * time.Minute

func loadApplicationVariables() (string, int, time.Duration) {
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

	return url, expectedStatusCode, timeout
}

func main() {
	url, expectedStatusCode, timeout := loadApplicationVariables()

	test := NewApplication(url, expectedStatusCode, timeout)
	appStatus := test.GetStatus()
	log.Println(appStatus)
}
