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
	client := createClient(test.Timeout)

	var resp *http.Response
	var err error
	var actualContent string
	var statusContentOk = true

	if test.ExpectedContent != "" {
		resp, err, actualContent, statusContentOk = performGetRequest(test, client)
		if err != nil {
			return createApplicationStatus(test, false, false, resp, err, true, "")
		}
	} else {
		resp, err = performHeadRequest(test, client)
		if err != nil {
			return createApplicationStatus(test, false, false, resp, err, false, "")
		}
	}

	statusOk := false
	if err == nil {
		statusOk = compareStatusCodes(resp.StatusCode, test.ExpectedStatusCode) &&
			compareLocations(resp.Header.Get("Location"), test.ExpectedLocation)
	}

	return createApplicationStatus(test, statusOk, statusContentOk, resp, nil, true, actualContent)
}

func createClient(timeout time.Duration) *http.Client {
	return &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Timeout: timeout,
	}
}

func getClientUrl(test Application) string {
	if test.ExpectedLocation != "" {
		return test.ExpectedLocation
	}

	return test.URL
}

func closeResponseBody(Body io.ReadCloser) {
	err := Body.Close()
	if err != nil {
		log.Println("Error closing response body:", err)
	}
}

func performGetRequest(test Application, client *http.Client) (*http.Response, error, string, bool) {
	clientUrl := getClientUrl(test)
	resp, err := client.Get(clientUrl)
	if err != nil {
		return nil, err, "", false
	}

	defer closeResponseBody(resp.Body)

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, resp.Body)
	if err != nil {
		log.Println("Error copying response body:", err)
	}

	actualContent := buf.String()
	statusContentOk, matchedContent := compareContent(actualContent, test.ExpectedContent)
	if statusContentOk {
		actualContent = matchedContent
	}

	return resp, nil, actualContent, statusContentOk
}

func performHeadRequest(test Application, client *http.Client) (*http.Response, error) {
	resp, err := client.Head(test.URL)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func createApplicationStatus(test Application, statusOk bool, statusContentOk bool, resp *http.Response, err error, isGet bool, actualContent string) *ApplicationStatus {
	actualStatusCode := 0
	actualLocation := ""

	if err != nil {
		log.Println("Error performing request:", err)
		statusOk = false
		statusContentOk = false
		actualContent = ""
	} else if resp != nil {
		actualStatusCode = resp.StatusCode
		actualLocation = resp.Header.Get("Location")
		if !isGet {
			actualContent = ""
		}
	}
	return &ApplicationStatus{
		Application:      &test,
		StatusOk:         statusOk,
		StatusContentOk:  statusContentOk,
		ActualStatusCode: actualStatusCode,
		ActualLocation:   actualLocation,
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
	if results.ActualStatusCode != results.Application.ExpectedStatusCode && results.ActualLocation != results.Application.ExpectedLocation {
		return fmt.Sprintf("Failure: URL %s resolved with %d, expected %d, and redirect location %s did not match %s", results.Application.URL, results.ActualStatusCode, results.Application.ExpectedStatusCode, results.ActualLocation, results.Application.ExpectedLocation)
	} else if results.ActualLocation != results.Application.ExpectedLocation && results.ActualStatusCode == results.Application.ExpectedStatusCode {
		return fmt.Sprintf("Failure: URL %s resolved with %d, but redirect location %s did not match %s", results.Application.URL, results.ActualStatusCode, results.ActualLocation, results.Application.ExpectedLocation)
	} else if results.ActualStatusCode != results.Application.ExpectedStatusCode && results.ActualLocation == results.Application.ExpectedLocation {
		return fmt.Sprintf("Failure: URL %s resolved with %d, expected %d, but redirect location matched", results.Application.URL, results.ActualStatusCode, results.Application.ExpectedStatusCode)
	} else if results.ActualStatusCode != results.Application.ExpectedStatusCode {
		return fmt.Sprintf("Failure: URL %s resolved with %d, expected %d", results.Application.URL, results.ActualStatusCode, results.Application.ExpectedStatusCode)
	} else {
		// Should not be reached under normal circumstances
		return fmt.Sprintf("Unknown failure for URL %s", results.Application.URL)
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
		if results.Application.Name == "alerts" {
			return fmt.Sprintf("Failure: Expected content %s did not match ActualContent for application 'alerts'", results.Application.ExpectedContent)
		} else {
			return fmt.Sprintf("Failure: Expected content %s did not match ActualContent %s", results.Application.ExpectedContent, results.ActualContent)
		}
	} else {
		return fmt.Sprintf("No content to compare")
	}
}
