package application

import (
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGetStatus(t *testing.T) {
	// Create mock server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/html":
			_, _ = fmt.Fprint(w, "<html><body><h1>Herman Melville</h1></body></html>")
		case "/notfound":
			w.WriteHeader(http.StatusNotFound)
		case "/timeout":
			time.Sleep(200 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
		case "/wrongstatus":
			w.WriteHeader(http.StatusNotFound)
		case "/500":
			w.WriteHeader(http.StatusInternalServerError)
		case "/slowresponse":
			time.Sleep(200 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer mockServer.Close()

	var tests = []struct {
		description              string
		application              *Application
		expectedSuccess          bool
		expectedActualStatusCode int
		expectedActualLocation   string
		expectedContentSuccess   bool
		expectedActualContent    string
	}{
		{"Success: correct redirect expected", &Application{"", "http://library.nyu.edu", http.StatusMovedPermanently, 800 * time.Millisecond, "https://library.nyu.edu/", ""}, true, http.StatusMovedPermanently, "https://library.nyu.edu/", true, ""},
		{"Failure: wrong redirect expected", &Application{"", "http://library.nyu.edu", http.StatusFound, 800 * time.Millisecond, "", ""}, false, http.StatusMovedPermanently, "", true, ""},
		{"Success: correct error expected", &Application{"", "https://library.nyu.edu/nopageexistshere", http.StatusNotFound, 600 * time.Millisecond, "", ""}, true, http.StatusNotFound, "", true, ""},
		{"Success: success status code expected", &Application{"", "https://library.nyu.edu", http.StatusOK, 800 * time.Millisecond, "", ""}, true, http.StatusOK, "", true, ""},
		{"Failure: wrong status code expected", &Application{"", mockServer.URL + "/wrongstatus", http.StatusOK, 800 * time.Millisecond, "", ""}, false, http.StatusNotFound, "", true, ""},
		{"Failure: application is down", &Application{"", mockServer.URL + "/500", http.StatusOK, 800 * time.Millisecond, "", ""}, false, http.StatusInternalServerError, "", true, ""},
		{"Success: timeout", &Application{"", "https://library.nyu.edu", http.StatusOK, 200 * time.Millisecond, "", ""}, true, http.StatusOK, "", true, ""},
		{"Failure: timeout", &Application{"", mockServer.URL + "/slowresponse", http.StatusOK, 1 * time.Millisecond, "", ""}, false, 0, "", false, ""},
		{"Success: correct redirect expected", &Application{"", "http://library.nyu.edu", http.StatusMovedPermanently, 800 * time.Millisecond, "https://library.nyu.edu/", ""}, true, http.StatusMovedPermanently, "https://library.nyu.edu/", true, ""},
		{"Failure: wrong redirect expected", &Application{"", "http://library.nyu.edu", http.StatusMovedPermanently, 800 * time.Millisecond, "http://library.nyu.edu/", ""}, false, http.StatusMovedPermanently, "https://library.nyu.edu/", true, ""},
		{"Failure: wrong redirect location expected", &Application{"", "http://library.nyu.edu", http.StatusMovedPermanently, 800 * time.Millisecond, "http://library.nyu.edu/", ""}, false, http.StatusMovedPermanently, "https://library.nyu.edu/", true, ""},
		{"Failure: wrong error expected", &Application{"", "https://library.nyu.edu/nopageexistshere", http.StatusFound, 800 * time.Millisecond, "", ""}, false, http.StatusNotFound, "", true, ""},
		{"Success: expected content found", &Application{"", mockServer.URL + "/html", http.StatusOK, 5 * time.Second, "", "Herman Melville"}, true, http.StatusOK, "", true, "Herman Melville"},
		{"Failure: expected content not found", &Application{"", mockServer.URL + "/html", http.StatusOK, 5 * time.Second, "", "Jules Verne - 20,000 Leagues Under the Sea"}, true, http.StatusOK, "", false, ""},
	}

	for _, test := range tests {
		t.Run(test.description, testGetStatusFunc(test.application, test.expectedSuccess, test.expectedActualStatusCode, test.expectedContentSuccess, test.expectedActualContent))
	}
}

func testGetStatusFunc(application *Application, expectedSuccess bool, expectedActualStatusCode int, expectedContentSuccess bool, actualContent string) func(*testing.T) {
	return func(t *testing.T) {
		status := application.GetStatus()
		assert.Equal(t, expectedSuccess, status.StatusOk)
		assert.Equal(t, expectedActualStatusCode, status.ActualStatusCode)
		assert.Equal(t, expectedContentSuccess, status.StatusContentOk)
		if expectedContentSuccess {
			assert.Contains(t, status.ActualContent, actualContent)
		}
	}
}

// This is a custom type that implements the http.RoundTripper interface: https://pkg.go.dev/net/http#RoundTripper
// It will be used to simulate a network error for your tests.
type errorTransport struct{}

// The RoundTrip method is what's called by http.Client when it makes a request.
// By always returning an error here, we can simulate network issues.
func (t *errorTransport) RoundTrip(_ *http.Request) (*http.Response, error) {
	return nil, errors.New("simulated network error")
}

func TestCreateApplicationStatus(t *testing.T) {
	// Create mock server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/successful":
			if r.Method == http.MethodGet {
				_, _ = fmt.Fprint(w, "Successful Request")
			}
		case "/failed":
			w.WriteHeader(http.StatusNotFound)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))
	defer mockServer.Close()

	// Define an application for testing
	app := Application{
		URL:                mockServer.URL + "/successful",
		ExpectedStatusCode: http.StatusOK,
		Timeout:            500 * time.Millisecond,
		ExpectedLocation:   "",
		ExpectedContent:    "Successful Request",
	}

	appWithoutContent := Application{
		URL:                mockServer.URL + "/successful",
		ExpectedStatusCode: http.StatusOK,
		Timeout:            500 * time.Millisecond,
		ExpectedLocation:   "",
		ExpectedContent:    "",
	}

	tests := []struct {
		description         string
		app                 Application
		statusOk            bool
		statusContentOk     bool
		expectedApplication *ApplicationStatus
	}{
		{
			description:     "GET request with network error",
			app:             app,
			statusOk:        false,
			statusContentOk: false,
			expectedApplication: &ApplicationStatus{
				Application:      &app,
				StatusOk:         false,
				StatusContentOk:  false,
				ActualStatusCode: 0,
				ActualLocation:   "",
				ActualContent:    "",
			},
		},
		{
			description:     "HEAD request with network error",
			app:             app,
			statusOk:        false,
			statusContentOk: false,
			expectedApplication: &ApplicationStatus{
				Application:      &app,
				StatusOk:         false,
				StatusContentOk:  false,
				ActualStatusCode: 0,
				ActualLocation:   "",
				ActualContent:    "",
			},
		},
		{
			description:     "Successful GET request",
			app:             app,
			statusOk:        true,
			statusContentOk: true,
			expectedApplication: &ApplicationStatus{
				Application:      &app,
				StatusOk:         true,
				StatusContentOk:  true,
				ActualStatusCode: http.StatusOK,
				ActualLocation:   "",
				ActualContent:    "Successful Request",
			},
		},
		{
			description:     "Successful HEAD request",
			app:             appWithoutContent,
			statusOk:        true,
			statusContentOk: true,
			expectedApplication: &ApplicationStatus{
				Application:      &appWithoutContent,
				StatusOk:         true,
				StatusContentOk:  true,
				ActualStatusCode: http.StatusOK,
				ActualLocation:   "",
				ActualContent:    "",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			// Here we create the http.Client to use in the tests.
			// If the test is supposed to simulate a network error, use the errorTransport defined above.
			// Otherwise, use a regular http.Client.
			var client *http.Client
			if test.description == "GET request with network error" || test.description == "HEAD request with network error" {
				client = &http.Client{
					Transport: &errorTransport{},
					Timeout:   test.app.Timeout,
				}
			} else {
				client = &http.Client{
					Timeout: test.app.Timeout,
				}
			}

			var resp *http.Response
			var err error
			var actualContent string

			if test.app.IsGet() {
				resp, err, actualContent, _ = performGetRequest(test.app, client)
			} else {
				resp, err = performHeadRequest(test.app, client)
			}

			result := createApplicationStatus(test.app, resp, err, actualContent)
			assert.Equal(t, test.expectedApplication, result)
		})
	}
}

func TestString(t *testing.T) {
	var tests = []struct {
		description    string
		appStatus      *ApplicationStatus
		expectedOutput string
	}{
		{description: "Successful status", appStatus: &ApplicationStatus{Application: &Application{"", "https://library.nyu.edu", http.StatusOK, time.Second, "", ""}, StatusOk: true, StatusContentOk: true, ActualStatusCode: 200}, expectedOutput: "Success: URL https://library.nyu.edu resolved with 200"},
		{description: "Failed status", appStatus: &ApplicationStatus{Application: &Application{URL: "https://library.nyu.edu", ExpectedStatusCode: http.StatusOK, Timeout: time.Second}, StatusOk: false, StatusContentOk: true, ActualStatusCode: 404}, expectedOutput: "Failure: URL https://library.nyu.edu resolved with 404, expected 200"},
		{description: "Successful status with location", appStatus: &ApplicationStatus{Application: &Application{URL: "http://library.nyu.edu", ExpectedStatusCode: http.StatusMovedPermanently, Timeout: time.Second, ExpectedLocation: "https://library.nyu.edu/"}, StatusOk: true, StatusContentOk: true, ActualStatusCode: 301, ActualLocation: "https://library.nyu.edu/"}, expectedOutput: "Success: URL http://library.nyu.edu resolved with 301, redirect location matched https://library.nyu.edu/"},
		{description: "Failed status with location", appStatus: &ApplicationStatus{Application: &Application{URL: "http://library.nyu.edu", ExpectedStatusCode: http.StatusMovedPermanently, Timeout: time.Second, ExpectedLocation: "http://library.nyu.edu/"}, StatusOk: false, StatusContentOk: true, ActualStatusCode: 301, ActualLocation: "https://library.nyu.edu/"}, expectedOutput: "Failure: URL http://library.nyu.edu resolved with 301, but redirect location https://library.nyu.edu/ did not match http://library.nyu.edu/"},
		{description: "Successful status with expected content", appStatus: &ApplicationStatus{Application: &Application{"", "https://example.com", http.StatusOK, time.Second, "", "Example Domain"}, StatusOk: true, StatusContentOk: true, ActualStatusCode: 200, ActualContent: "Example Domain"}, expectedOutput: "Success: URL https://example.com resolved with 200\nSuccess: ExpectedContent Example Domain matched ActualContent Example Domain"},
		{description: "Failed status with unexpected content", appStatus: &ApplicationStatus{Application: &Application{"", "https://example.com", http.StatusOK, time.Second, "", "Wrong Content"}, StatusOk: true, StatusContentOk: false, ActualStatusCode: 200, ActualContent: "Example Domain"}, expectedOutput: "Success: URL https://example.com resolved with 200\nFailure: Expected content Wrong Content did not match ActualContent"},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			assert.Equal(t, test.expectedOutput, test.appStatus.String())
		})
	}
}
