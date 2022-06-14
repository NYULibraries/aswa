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
	Timeout            time.Duration `default:"1 * time.Minute"`
	ExpectedLocation   string        `yaml:"expected_location"`
}

// NewApplication returns a Application initialized with specified values
func NewApplication(name string, url string, expectedStatusCode int, timeout time.Duration, expectedLocation string) *Application {
	return &Application{name, url, expectedStatusCode, timeout, expectedLocation}
}

// ApplicationStatus represents the results of a synthetic test
type ApplicationStatus struct {
	Application      *Application
	Success          bool
	ActualStatusCode int
	ActualLocation   string `default:""`
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

	if resp.StatusCode != test.ExpectedStatusCode || (test.ExpectedLocation != "" && resp.Header.Get("Location") != test.ExpectedLocation) {
		return &ApplicationStatus{&test, false, resp.StatusCode, resp.Header.Get("Location")}
	}

	return &ApplicationStatus{&test, true, resp.StatusCode, resp.Header.Get("Location")}
}

// String outputs the application status as a single string
func (results ApplicationStatus) String() string {

	if results.Success && results.ActualLocation != "" {
		return fmt.Sprintf("Success: URL %s resolved with %d, redirect location matched %s", results.Application.URL, results.ActualStatusCode, results.ActualLocation)
	} else if results.Success {
		return fmt.Sprintf("Success: URL %s resolved with %d", results.Application.URL, results.ActualStatusCode)
	} else if !results.Success && results.ActualLocation != "" && results.ActualStatusCode == results.Application.ExpectedStatusCode {
		return fmt.Sprintf("Failure: URL %s resolved with %d, but redirect location %s did not match %s", results.Application.URL, results.ActualStatusCode, results.ActualLocation, results.Application.ExpectedLocation)
	} else {
		return fmt.Sprintf("Failure: URL %s resolved with %d, expected %d, %s", results.Application.URL, results.ActualStatusCode, results.Application.ExpectedStatusCode, results.ActualLocation)
	}
}