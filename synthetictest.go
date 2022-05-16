package main

import (
	"net/http"
	"time"
)

// Application represents a synthetic test on an external url to perform
type Application struct {
	url                string
	expectedStatusCode int
	timeout            time.Duration
}

// NewApplication returns a Application initialized with specified values
func NewApplication(url string, expectedStatusCode int, timeout time.Duration) *Application {
	return &Application{url, expectedStatusCode, timeout}
}

// ApplicationStatus represents the results of a synthetic test
type ApplicationStatus struct {
	success          bool
	actualStatusCode int
	test             *Application
}

// GetStatus performs an HTTP call for the given Application's url and returns the ApplicationStatus corresponding to those results
func (test Application) GetStatus() *ApplicationStatus {

	// TODO: code the actual http call here

	return &ApplicationStatus{true, http.StatusOK, &test}
}

func (results ApplicationStatus) String() string {

	// TODO: code the formatting of results here

	return "Hello world!"
}
