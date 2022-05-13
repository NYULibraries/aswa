package main

import (
	"net/http"
	"time"
)

// SyntheticTest represents a synthetic test on an external url to perform
type SyntheticTest struct {
	url            string
	expectedStatus int
	timeout        time.Duration
}

// NewSyntheticTest returns a SyntheticTest initialized with specified values
func NewSyntheticTest(url string, expectedStatus int, timeout time.Duration) *SyntheticTest {
	return &SyntheticTest{url, expectedStatus, timeout}
}

// SyntheticTestResults represents the results of a synthetic test
type SyntheticTestResults struct {
	success      bool
	actualStatus int
	test         *SyntheticTest
}

// GetResults performs an HTTP call for the given SyntheticTest's url and returns the SyntheticTestResults corresponding to those results
func (test SyntheticTest) GetResults() *SyntheticTestResults {

	// TODO: code the actual http call here

	return &SyntheticTestResults{true, http.StatusOK, &test}
}

func (results SyntheticTestResults) String() string {

	// TODO: code the formatting of results here

	return "Hello world!"
}
