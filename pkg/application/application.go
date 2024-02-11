package application

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

const envDebugMode = "DEBUG_MODE"

var (
	DebugMode = os.Getenv(envDebugMode) == "true"
	IsPrimoVE bool
)

// Application represents a synthetic test on an external url to perform
type Application struct {
	Name               string        `yaml:"name"`
	URL                string        `yaml:"url"`
	ExpectedStatusCode int           `yaml:"expected_status"`
	Timeout            time.Duration `yaml:"timeout"`
	ExpectedLocation   string        `yaml:"expected_location"`
	ExpectedContent    string        `yaml:"expected_content"`
	ExpectedCSP        string        `yaml:"expected_csp"`
}

// ApplicationStatus represents the results of a synthetic test
type ApplicationStatus struct {
	Application      *Application
	StatusOk         bool
	StatusContentOk  bool
	StatusCSPOk      bool
	ActualStatusCode int
	ActualLocation   string `default:""`
	ActualContent    string `default:""`
	ActualCSP        string `default:""`
}

// SetIsPrimoVE sets the IsPrimoVE flag based on the yamlPath.
func SetIsPrimoVE(yamlPath string) {
	IsPrimoVE = yamlPath == "config/primo_ve.applications.yml"
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
		return false, actual
	}
	// The slice actual[index : index+len(expected)] starts at the index where the expected string is found and ends at the index after the last character of the expected string.
	return true, actual[index : index+len(expected)]
}

// compareCSP compares the actual and expected CSP.
func compareCSP(actual string, expected string) bool {
	if expected == "" {
		return true // No CSP check required
	}
	return actual == expected
}

// GetStatus performs an HTTP call for the given Application's url, checks the expected status code, location, and content, and returns the ApplicationStatus corresponding to those results.
// If the expected content is not empty, the function will also perform a GET request to retrieve and compare the content.
func (test Application) GetStatus() *ApplicationStatus {
	client := createClient(test.Timeout)

	var resp *http.Response
	var err error
	var actualContent string

	if test.IsGet() {
		resp, err, actualContent, _ = performGetRequest(test, client)
		if err != nil {
			return createApplicationStatus(test, resp, err, "")
		}
	} else {
		resp, err = performHeadRequest(test, client)
		if err != nil {
			return createApplicationStatus(test, resp, err, "")
		}
	}

	return createApplicationStatus(test, resp, nil, actualContent)
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

func (test Application) IsGet() bool {
	return test.ExpectedContent != ""
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

func createApplicationStatus(test Application, resp *http.Response, err error, actualContent string) *ApplicationStatus {
	statusOk := false
	statusContentOk := true
	statusCSPOk := true
	actualStatusCode := 0
	actualLocation := ""
	actualCSP := ""

	if err != nil {
		log.Println("Error performing request:", err)
		actualContent = ""
		statusContentOk = false
		statusCSPOk = false
	} else if resp != nil {
		actualStatusCode = resp.StatusCode
		actualLocation = resp.Header.Get("Location")

		// Determine the statusOk
		statusOk = compareStatusCodes(resp.StatusCode, test.ExpectedStatusCode) &&
			compareLocations(actualLocation, test.ExpectedLocation)

		// Determine the statusContentOk
		if test.IsGet() {
			statusContentOk, actualContent = compareContent(actualContent, test.ExpectedContent)
		} else {
			actualContent = ""
		}
		// Determine the statusCSPOk
		if test.ExpectedCSP != "" {
			actualCSP = resp.Header.Get("Content-Security-Policy")
			statusCSPOk = compareCSP(actualCSP, test.ExpectedCSP)
		} else {
			actualCSP = ""
		}
	}

	return &ApplicationStatus{
		Application:      &test,
		StatusOk:         statusOk,
		StatusContentOk:  statusContentOk,
		StatusCSPOk:      statusCSPOk,
		ActualStatusCode: actualStatusCode,
		ActualLocation:   actualLocation,
		ActualContent:    actualContent,
		ActualCSP:        actualCSP,
	}
}

// String outputs the application status as a single string
func (results ApplicationStatus) String() string {
	statusString := ""
	contentString := ""
	cspString := ""

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

	// Handling the CSP check status
	if results.Application.ExpectedCSP != "" {
		if results.StatusCSPOk {
			cspString = cspSuccessString(results)
		} else {
			cspString = cspFailureString(results)
		}
	}

	if contentString != "" || cspString != "" {
		return statusString + "\n" + contentString + "\n" + cspString
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
	actualStatusCode := results.ActualStatusCode
	expectedStatusCode := results.Application.ExpectedStatusCode
	actualLocation := results.ActualLocation
	expectedLocation := results.Application.ExpectedLocation
	url := results.Application.URL

	statusMatch := actualStatusCode == expectedStatusCode
	locationMatch := actualLocation == expectedLocation

	var mismatchDetails string

	if expectedLocation != "" {
		if !statusMatch && !locationMatch {
			mismatchDetails = fmt.Sprintf("resolved with %d, expected %d, and redirect location %s did not match %s", actualStatusCode, expectedStatusCode, actualLocation, expectedLocation)
		} else if statusMatch && !locationMatch {
			mismatchDetails = fmt.Sprintf("resolved with %d, but redirect location %s did not match %s", actualStatusCode, actualLocation, expectedLocation)
		} else if !statusMatch && locationMatch {
			mismatchDetails = fmt.Sprintf("resolved with %d, expected %d, but redirect location matched", actualStatusCode, expectedStatusCode)
		}
	} else if !statusMatch {
		mismatchDetails = fmt.Sprintf("resolved with %d, expected %d", actualStatusCode, expectedStatusCode)
	} else {
		// Should not be reached under normal circumstances
		return fmt.Sprintf("Unknown failure for URL %s", url)
	}

	return fmt.Sprintf("Failure: URL %s %s", url, mismatchDetails)
}

func contentSuccessString(results ApplicationStatus) string {
	if results.ActualContent != "" {
		return fmt.Sprintf("Success: ExpectedContent %s matched ActualContent %s", results.Application.ExpectedContent, results.ActualContent)
	} else {
		return fmt.Sprintf("No content to compare")
	}
}

func contentFailureString(results ApplicationStatus) string {
	log.Printf("DebugMode: %t, IsPrimoVE: %t", DebugMode, IsPrimoVE)
	if results.ActualContent != "" {
		if results.Application.Name != "circleCI" {
			if IsPrimoVE && DebugMode {
				// For Primo VE checks with debug mode enabled, the actual content is included in the failure message
				return fmt.Sprintf("Failure: Expected content %s did not match Actual Content %s", results.Application.ExpectedContent, results.ActualContent)
			} else {
				return fmt.Sprintf("Failure: Expected content %s did not match Actual Content", results.Application.ExpectedContent)
			}
		} else {
			return fmt.Sprintf("Failure: Expected content %s did not match ActualContent %s", results.Application.ExpectedContent, results.ActualContent)
		}
	} else {
		return fmt.Sprintf("Failure: No content to compare")
	}
}

func cspSuccessString(results ApplicationStatus) string {
	if results.ActualCSP != "" {
		return fmt.Sprintf("Success: Expected Primo VE CSP header matched Actual CSP header")
	}
	return ""
}

func cspFailureString(results ApplicationStatus) string {
	if results.ActualCSP != "" {
		return fmt.Sprintf("Failure: Expected Primo VE CSP header did not match Actual CSP header: %s", results.ActualCSP)
	} else {
		return fmt.Sprintf("Failure: No Primo VE CSP header to compare")
	}
}
