package application

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	envDebugMode = "DEBUG_MODE"
	userAgent    = "ASWA-MonitoringService (HealthCheck; contact: lib-appdev@nyu.edu)"
)

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

// AppCheckStatus represents the results of a synthetic test
type AppCheckStatus struct {
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
	IsPrimoVE = yamlPath == "./config/primo_ve.applications.yml"
}

// compareStatusCodes compares the actual and expected status codes.
// It returns true if they are equal, and false otherwise.
func compareStatusCodes(actual, expected int) bool {
	return actual == expected
}

// compareLocations compares an actual redirect location against an expected one.
//
// Logic summary:
// 1. If either URL fails to parse, it falls back to a raw string comparison.
// 2. If the expected location is relative (no scheme/host), only the path and query are compared.
// 3. Otherwise (absolute expected URL), the full absolute URLs are compared.
//
// Returns true if the locations are considered equivalent according to the above rules.
func compareLocations(actualLocation, expectedLocation string) bool {
	parsedActualURL, actualParseErr := url.Parse(actualLocation)
	parsedExpectedURL, expectedParseErr := url.Parse(expectedLocation)

	// Fallback: if either parse fails, compare raw strings
	if actualParseErr != nil || expectedParseErr != nil {
		return actualLocation == expectedLocation
	}

	// If the expected location is relative (no scheme/host), compare only path + query
	if parsedExpectedURL.Scheme == "" && parsedExpectedURL.Host == "" {
		actualURI := parsedActualURL.RequestURI()
		expectedURI := parsedExpectedURL.RequestURI()

		// Accept relative paths with or without a leading slash
		if actualURI == expectedURI {
			return true
		}
		if "/"+expectedURI == actualURI {
			return true
		}
		if strings.TrimPrefix(actualURI, "/") == expectedURI {
			return true
		}

		return false
	}

	// Otherwise, compare the full absolute URLs
	return parsedActualURL.String() == parsedExpectedURL.String()
}

// compareContent compares the actual and expected content.
func compareContent(actual, expected string) (bool, string) {
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

// GetStatus performs an HTTP request for the given application's URL and evaluates
// its response against expected criteria such as status code, redirect location,
// and optional content or CSP header.
//
// Behavior summary:
//   - Always validates the original URL’s HTTP status and redirect (no auto-follow).
//   - If ExpectedContent is configured, also performs a GET request to fetch and
//     validate page content (optionally following the expected redirect).
func (test Application) GetStatus() *AppCheckStatus {
	client := createClient(test.Timeout)

	var resp *http.Response
	var err error
	var actualContent string
	var statusContentOk bool

	// Phase 1: probe ORIGINAL URL (status + Location)
	resp, err = performHeadRequest(test, client)
	if err != nil {
		return createApplicationStatus(test, resp, err, "", false)
	}
	if resp == nil {
		return createApplicationStatus(test, resp, fmt.Errorf("nil HEAD response"), "", false)
	}
	defer closeResponseBody(resp.Body)

	if DebugMode {
		log.Printf("[HEAD probe] url=%s status=%d location=%q",
			test.URL, resp.StatusCode, resp.Header.Get("Location"))
	}

	// Phase 2: content on the FINAL landing page (follow all redirects)
	if test.IsGet() {
		// Clone client and set redirect handler for visibility and cap
		followClient := *client
		followClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			if DebugMode {
				prev := via[len(via)-1].URL
				log.Printf("[GET redirect] hop=%d %s -> %s", len(via), prev, req.URL)
			}
			if len(via) >= 10 {
				return fmt.Errorf("stopped after %d redirects", len(via))
			}
			return nil
		}

		contentApp := test
		contentApp.URL = test.URL

		if DebugMode {
			log.Printf("[GET start] url=%s", contentApp.URL)
		}

		var respStatusCode int
		var finalURL string
		respStatusCode, finalURL, actualContent, statusContentOk, err =
			performGetRequest(contentApp, &followClient)
		if err != nil {
			if DebugMode {
				log.Printf("[GET error] url=%s error=%v", contentApp.URL, err)
			}
			return createApplicationStatus(test, nil, err, "", false)
		}

		if DebugMode {
			log.Printf("[GET final] status=%d url=%s bodyLen=%d",
				respStatusCode, finalURL, len(actualContent))
		}
	} else {
		statusContentOk = true
	}

	return createApplicationStatus(test, resp, nil, actualContent, statusContentOk)
}

func createClient(timeout time.Duration) *http.Client {
	return &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Timeout: timeout,
	}
}

func (test Application) IsGet() bool {
	return test.ExpectedContent != ""
}

func closeResponseBody(body io.ReadCloser) {
	if err := body.Close(); err != nil {
		log.Println("Error closing response body:", err)
	}
}

func performGetRequest(test Application, client *http.Client) (int, string, string, bool, error) {
	req, err := http.NewRequest(http.MethodGet, test.URL, nil)
	if err != nil {
		return 0, "", "", false, err
	}

	req.Header.Set("User-Agent", userAgent)

	resp, err := client.Do(req)
	if err != nil {
		return 0, "", "", false, err
	}
	defer closeResponseBody(resp.Body)

	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, resp.Body); err != nil {
		log.Println("Error copying response body:", err)
		return resp.StatusCode, resp.Request.URL.String(), "", false, err
	}

	actualContent := buf.String()
	statusContentOk, matchedContent := compareContent(actualContent, test.ExpectedContent)
	if statusContentOk {
		actualContent = matchedContent
	}

	finalURL := ""
	if resp.Request != nil && resp.Request.URL != nil {
		finalURL = resp.Request.URL.String()
	}

	return resp.StatusCode, finalURL, actualContent, statusContentOk, nil
}

func performHeadRequest(test Application, client *http.Client) (*http.Response, error) {
	resp, err := client.Head(test.URL)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func createApplicationStatus(test Application, resp *http.Response, err error, actualContent string, statusContentOk bool) *AppCheckStatus {
	statusOk := false
	statusCSPOk := true
	actualStatusCode := 0
	actualLocation := ""
	actualCSP := ""

	if err != nil {
		log.Printf("[%s] Request error: %v", test.Name, err)
		actualContent = ""
		statusContentOk = false
		statusCSPOk = false
	} else if resp != nil {
		actualStatusCode = resp.StatusCode
		actualLocation = resp.Header.Get("Location")

		// Determine the statusOk
		statusOk = compareStatusCodes(resp.StatusCode, test.ExpectedStatusCode) &&
			(test.ExpectedLocation == "" || compareLocations(actualLocation, test.ExpectedLocation))

		if !test.IsGet() {
			actualContent = ""
		}
		// Determine the statusCSPOk
		if test.ExpectedCSP != "" {
			actualCSP = resp.Header.Get("Content-Security-Policy")
			statusCSPOk = compareCSP(actualCSP, test.ExpectedCSP)
		}
	}

	return &AppCheckStatus{
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
func (results AppCheckStatus) String() string {
	var output []string

	if results.StatusOk {
		output = append(output, successString(results))
	} else {
		output = append(output, failureString(results))
	}

	if results.Application.ExpectedContent != "" {
		if results.StatusContentOk {
			output = append(output, contentSuccessString(results))
		} else {
			output = append(output, contentFailureString(results))
		}
	}

	// Handling the CSP check status
	if results.Application.ExpectedCSP != "" {
		if results.StatusCSPOk {
			output = append(output, cspSuccessString(results))
		} else {
			output = append(output, cspFailureString(results))
		}
	}

	return strings.Join(output, "\n")
}

func successString(results AppCheckStatus) string {
	if results.ActualLocation != "" {
		return fmt.Sprintf("Success: URL %s resolved with %d, redirect location matched %s", results.Application.URL, results.ActualStatusCode, results.ActualLocation)
	}
	return fmt.Sprintf("Success: URL %s resolved with %d", results.Application.URL, results.ActualStatusCode)

}

func failureString(results AppCheckStatus) string {
	actualStatusCode := results.ActualStatusCode
	expectedStatusCode := results.Application.ExpectedStatusCode
	actualLocation := results.ActualLocation
	expectedLocation := results.Application.ExpectedLocation
	appURL := results.Application.URL

	statusMatch := compareStatusCodes(actualStatusCode, expectedStatusCode)
	locationMatch := compareLocations(actualLocation, expectedLocation)

	var mismatchDetails string

	if expectedLocation != "" {
		if !statusMatch && !locationMatch {
			mismatchDetails = fmt.Sprintf("resolved with %d, expected %d, and redirect location %s did not match %s", actualStatusCode, expectedStatusCode, actualLocation, expectedLocation)
		} else if statusMatch && !locationMatch {
			mismatchDetails = fmt.Sprintf("resolved with %d, but redirect location %s did not match %s", actualStatusCode, actualLocation, expectedLocation)
		} else if !statusMatch {
			mismatchDetails = fmt.Sprintf("resolved with %d, expected %d, but redirect location matched", actualStatusCode, expectedStatusCode)
		}
	} else if !statusMatch {
		mismatchDetails = fmt.Sprintf("resolved with %d, expected %d", actualStatusCode, expectedStatusCode)
	} else {
		// Should not be reached under normal circumstances
		return fmt.Sprintf("Unknown failure for URL %s", appURL)
	}

	return fmt.Sprintf("Failure: URL %s %s", appURL, mismatchDetails)
}

func contentSuccessString(results AppCheckStatus) string {
	if results.ActualContent != "" {
		return fmt.Sprintf("Success: ExpectedContent %s matched ActualContent %s", results.Application.ExpectedContent, results.ActualContent)
	}
	return "No content to compare"

}

func contentFailureString(results AppCheckStatus) string {
	log.Printf("DebugMode: %t, IsPrimoVE: %t", DebugMode, IsPrimoVE)
	if results.ActualContent == "" {
		return "Failure: No content to compare"
	}

	if results.Application.Name == "circleCI" {
		return fmt.Sprintf("Failure: Expected content %s did not match ActualContent %s", results.Application.ExpectedContent, results.ActualContent)
	}

	if IsPrimoVE && DebugMode {
		// For Primo VE checks with debug mode enabled, the actual content is included in the failure message
		return fmt.Sprintf("Failure: Expected content %s did not match Actual Content %s", results.Application.ExpectedContent, results.ActualContent)
	}

	return fmt.Sprintf("Failure: Expected content %s did not match Actual Content", results.Application.ExpectedContent)
}

func cspSuccessString(results AppCheckStatus) string {
	if results.ActualCSP != "" {
		return "Success: Expected Primo VE CSP header matched Actual CSP header"
	}
	return ""
}

func cspFailureString(results AppCheckStatus) string {
	if results.ActualCSP != "" {
		return fmt.Sprintf("Failure: Expected Primo VE CSP header did not match Actual CSP header: %s", results.ActualCSP)
	}
	return "Failure: No Primo VE CSP header to compare"
}
