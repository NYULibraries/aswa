package application

import (
	"fmt"
	"net/http"
	"time"
)

// Application represents a synthetic test on an external url to perform
type Application struct {
	Name               string        `yaml:"name"`
	URL                string        `yaml:"url"`
	ExpectedStatusCode int           `yaml:"expected_status"`
	Timeout            time.Duration `yaml:"timeout"`
	ExpectedLocation   string        `yaml:"expected_location"`
	ExpectedCDN        string        `yaml:"expected_cdn"`
}

// ApplicationStatus represents the results of a synthetic test
type ApplicationStatus struct {
	Application      *Application
	Success          bool
	ActualStatusCode int
	ActualLocation   string `default:""`
}

// compareStatusCodes compares the actual and expected status codes.
// It returns true if they are equal, and false otherwise.
func compareStatusCodes(actual int, expected int) bool {
	return actual == expected
}

// compareLocations compares the actual and expected locations.
// It returns true if they are equal, and false otherwise.
func compareLocations(actual string, expected string) bool {
	return actual == expected
}

// GetStatus performs an HTTP call for the given Application's url and returns the ApplicationStatus corresponding to those results
func (test Application) GetStatus() *ApplicationStatus {

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Timeout: test.Timeout,
	}

	resp, err := client.Head(test.URL)
	if err != nil {
		return &ApplicationStatus{&test, false, 0, ""}
	}

	statusOk := compareStatusCodes(resp.StatusCode, test.ExpectedStatusCode) && compareLocations(resp.Header.Get("Location"), test.ExpectedLocation)
	//statusOk := resp.StatusCode == test.ExpectedStatusCode && resp.Header.Get("Location") == test.ExpectedLocation
	return &ApplicationStatus{&test, statusOk, resp.StatusCode, resp.Header.Get("Location")}
}

// String outputs the application status as a single string
func (results ApplicationStatus) String() string {
	if results.Success {
		return successString(results)
	} else {
		return failureString(results)
	}
}

func successString(results ApplicationStatus) string {
	if results.ActualLocation != "" {
		return fmt.Sprintf("Success: URL %s resolved with %d, redirect location matched %s", results.Application.URL, results.ActualStatusCode, results.ActualLocation)
	} else {
		return fmt.Sprintf("Success: URL %s resolved with %d", results.Application.URL, results.ActualStatusCode)
	}
}

func failureString(results ApplicationStatus) string {
	if results.ActualLocation != "" && results.ActualStatusCode == results.Application.ExpectedStatusCode {
		return fmt.Sprintf("Failure: URL %s resolved with %d, but redirect location %s did not match %s", results.Application.URL, results.ActualStatusCode, results.ActualLocation, results.Application.ExpectedLocation)
	} else {
		return fmt.Sprintf("Failure: URL %s resolved with %d, expected %d, %s", results.Application.URL, results.ActualStatusCode, results.Application.ExpectedStatusCode, results.ActualLocation)
	}
}
