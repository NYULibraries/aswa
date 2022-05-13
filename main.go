package main

import (
	"log"
	"os"
	"strconv"
	"time"
)

const defaultTimeout = 1 * time.Minute

func main() {
	url := os.Getenv("ASWA_URL")
	expectedStatus, err := strconv.Atoi(os.Getenv("ASWA_EXPECTED_STATUS"))
	if err != nil {
		log.Println("Could not parse expected status; aborting!")
		panic(err)
	}

	timeout, err := time.ParseDuration(os.Getenv("ASWA_TIMEOUT"))
	if timeout <= 0 || err != nil {
		log.Println("No valid timeout provided; using default")
		timeout = defaultTimeout
	}

	test := NewSyntheticTest(url, expectedStatus, timeout)
	results := test.GetResults()
	log.Println(results)
}
