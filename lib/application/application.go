package application

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

// Application represents a synthetic test on an external url to perform
type Application struct {
	Name               string        `yaml:"name"`
	URL                string        `yaml:"url"`
	ExpectedStatusCode int           `yaml:"expected_status"`
	Timeout            time.Duration `yaml:"timeout"`
	ExpectedLocation   string        `yaml:"expected_location"`
	ExpectedContent    string        `yaml:"expected_content"`
}

// ApplicationStatus represents the results of a synthetic test
type ApplicationStatus struct {
	Application      *Application
	StatusOk         bool
	StatusContentOk  bool
	ActualStatusCode int
	ActualLocation   string `default:""`
	ActualContent    string `default:""`
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

// compareContent compares the actual and expected content.
func compareContent(actual string, expected string) (bool, string) {
	index := strings.Index(actual, expected)
	if index == -1 {
		return false, ""
	}
	// The slice actual[index : index+len(expected)] starts at the index where the expected string is found and ends at the index after the last character of the expected string.
	return true, actual[index : index+len(expected)]
}

// GetStatus performs an HTTP call for the given Application's url, checks the expected status code, location, and content, and returns the ApplicationStatus corresponding to those results.
// If the expected content is not empty, the function will also perform a GET request to retrieve and compare the content.
func (test Application) GetStatus() *ApplicationStatus {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Timeout: test.Timeout,
	}

	respHead, err := client.Head(test.URL)
	if err != nil {
		log.Println("Error performing HEAD request:", err)
		return &ApplicationStatus{
			Application:      &test,
			StatusOk:         false,
			StatusContentOk:  false,
			ActualStatusCode: 0,
			ActualLocation:   "",
			ActualContent:    "",
		}
	}

	statusOk := compareStatusCodes(respHead.StatusCode, test.ExpectedStatusCode) &&
		compareLocations(respHead.Header.Get("Location"), test.ExpectedLocation)

	var actualContent string
	var matchedContent string
	var statusContentOk bool

	if test.ExpectedContent != "" {
		var clientUrl string

		if test.ExpectedLocation != "" {
			clientUrl = test.ExpectedLocation
		} else {
			clientUrl = test.URL
		}

		respGet, err := client.Get(clientUrl)
		if err != nil {
			log.Println("Error performing GET request:", err)
			return &ApplicationStatus{
				Application:      &test,
				StatusOk:         statusOk,
				StatusContentOk:  false,
				ActualStatusCode: respHead.StatusCode,
				ActualLocation:   respHead.Header.Get("Location"),
				ActualContent:    "",
			}
		}
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				log.Println("Error closing response body:", err)
			}
		}(respGet.Body)

		buf := new(bytes.Buffer)
		_, err = io.Copy(buf, respGet.Body)
		if err != nil {
			log.Println("Error copying response body:", err)
		}

		actualContent = buf.String()
		statusContentOk, matchedContent = compareContent(actualContent, test.ExpectedContent)
		if statusContentOk {
			actualContent = matchedContent
		}
	} else {
		statusContentOk = true
	}

	return &ApplicationStatus{
		Application:      &test,
		StatusOk:         statusOk,
		StatusContentOk:  statusContentOk,
		ActualStatusCode: respHead.StatusCode,
		ActualLocation:   respHead.Header.Get("Location"),
		ActualContent:    actualContent,
	}
}

// String outputs the application status as a single string
func (results ApplicationStatus) String() string {
	statusString := ""
	contentString := ""

	if results.StatusOk {
		statusString = successString(results)
	} else {
		statusString = failureString(results)
	}

	if results.Application.ExpectedContent != "" {
		if results.StatusContentOk {
			contentString = contentSuccessString(results)
		} else {
			contentString = contentFailureString(results)
		}
	}

	if contentString != "" {
		return statusString + "\n" + contentString
	} else {
		return statusString
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

func contentSuccessString(results ApplicationStatus) string {
	if results.ActualContent != "" {
		return fmt.Sprintf("Success: ExpectedContent %s matched ActualContent %s", results.Application.ExpectedContent, results.ActualContent)
	} else {
		return fmt.Sprintf("No content to compare")
	}
}

func contentFailureString(results ApplicationStatus) string {
	if results.ActualContent != "" {
		return fmt.Sprintf("Failure: Expected content %s did not match ActualContent %s", results.Application.ExpectedContent, results.ActualContent)
	} else {
		return fmt.Sprintf("No content to compare")
	}
}
